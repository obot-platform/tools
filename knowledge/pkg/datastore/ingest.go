package datastore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings"
	"github.com/gptscript-ai/knowledge/pkg/index/types"
	"github.com/gptscript-ai/knowledge/pkg/log"
	"github.com/gptscript-ai/knowledge/pkg/output"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	cg "github.com/philippgille/chromem-go"

	"github.com/google/uuid"
	"github.com/gptscript-ai/knowledge/pkg/datastore/filetypes"
	"github.com/gptscript-ai/knowledge/pkg/datastore/transformers"
	"github.com/gptscript-ai/knowledge/pkg/flows"
)

type IngestOpts struct {
	FileMetadata        *types.FileMetadata
	IsDuplicateFuncName string
	IsDuplicateFunc     IsDuplicateFunc
	IngestionFlows      []flows.IngestionFlow
	ExtraMetadata       map[string]any
	ReuseEmbeddings     bool
	ReuseFiles          bool
}

// Ingest loads a document from a reader and adds it to the dataset.
func (s *Datastore) Ingest(ctx context.Context, datasetID string, filename string, content []byte, opts IngestOpts) ([]string, error) {
	ingestionStart := time.Now()
	if filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	statusLog := log.FromCtx(ctx).With("phase", "store")

	// Get dataset
	ds, err := s.GetDataset(ctx, datasetID, nil)
	if err != nil {
		return nil, err
	}

	// Dataset does not exist - create it if requested, else error out
	if ds == nil {
		return nil, fmt.Errorf("dataset %q not found", datasetID)
	}

	// Check if Dataset has an embedding config attached
	if ds.EmbeddingsProviderConfig == nil {
		slog.Debug("Embeddingsconfig", "config", s.EmbeddingConfig)
		ncfg, err := embeddings.AsEmbeddingModelProviderConfig(s.EmbeddingModelProvider, true)
		if err != nil {
			return nil, fmt.Errorf("failed to get embedding model provider config: %w", err)
		}
		nds := types.Dataset{
			ID:                       datasetID,
			EmbeddingsProviderConfig: &ncfg,
		}
		ds, err = s.UpdateDataset(ctx, nds, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to update dataset: %w", err)
		}
	}
	if ds.EmbeddingsProviderConfig != nil {
		if s.EmbeddingModelProvider.Name() != ds.EmbeddingsProviderConfig.Type {
			slog.Warn("Embeddings provider mismatch", "dataset", datasetID, "attached", ds.EmbeddingsProviderConfig.Type, "configured", s.EmbeddingModelProvider.Name())
		}

		dsEmbeddingProvider, err := embeddings.ProviderFromConfig(*ds.EmbeddingsProviderConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get embeddings model provider: %w", err)
		}

		if s.EmbeddingModelProvider.EmbeddingModelName() != dsEmbeddingProvider.EmbeddingModelName() {
			slog.Warn("Embeddings model mismatch", "dataset", datasetID, "attached", dsEmbeddingProvider.EmbeddingModelName(), "configured", s.EmbeddingModelProvider.EmbeddingModelName())
			if os.Getenv("KNOW_PREFER_NEW_EMBEDDING_MODEL") == "" {
				slog.Info("Using dataset's embeddings model", "model", dsEmbeddingProvider.EmbeddingModelName())
				s.EmbeddingModelProvider.UseEmbeddingModel(dsEmbeddingProvider.EmbeddingModelName())
			}
		}

		if os.Getenv("KNOW_STRICT_EMBEDDING_CONFIG_CHECK") != "" {
			err = embeddings.CompareRequiredFields(s.EmbeddingModelProvider.Config(), dsEmbeddingProvider.Config())
			if err != nil {
				slog.Info("Dataset has attached embeddings provider config", "config", output.RedactSensitive(ds.EmbeddingsProviderConfig))
				return nil, fmt.Errorf("mismatching embedding provider configs: %w", err)
			}
		}
	}

	// File Deduplication
	isDuplicate := DedupeUpsert // default: no deduplication
	if opts.IsDuplicateFuncName != "" {
		df, ok := IsDuplicateFuncs[opts.IsDuplicateFuncName]
		if !ok {
			return nil, fmt.Errorf("unknown deduplication function: %s", opts.IsDuplicateFuncName)
		}
		isDuplicate = df
	} else if opts.IsDuplicateFunc != nil {
		isDuplicate = opts.IsDuplicateFunc
	}

	// Generate ID
	fUUID, err := uuid.NewUUID()
	if err != nil {
		slog.Error("Failed to generate UUID", "error", err)
		return nil, err
	}
	fileID := fUUID.String()

	/*
	 * Detect filetype
	 */

	filetype, err := filetypes.GetFiletype(filename, content)
	if err != nil {
		return nil, err
	}

	statusLog = statusLog.With("filename", filename, "filetype", filetype)

	slog.Debug("Loading data", "type", filetype, "filename", filename, "size", len(content))

	/*
	 * Exit early if the document is a duplicate
	 */
	isDupe, err := isDuplicate(ctx, s, datasetID, nil, opts)
	if err != nil {
		statusLog.With("status", "failed").Error("Failed to check for duplicates", "error", err)
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}
	if isDupe {
		statusLog.With("status", "skipped").With("reason", "duplicate").Info("Ignoring duplicate document")
		return nil, nil
	}

	/*
	 * Load the ingestion flow - custom or default config or mixture of both
	 */
	ingestionFlow := flows.IngestionFlow{}
	for _, flow := range opts.IngestionFlows {
		if flow.SupportsFiletype(filetype) {
			ingestionFlow = flow
			break
		}
	}

	if err := ingestionFlow.FillDefaults(filetype); err != nil {
		return nil, err
	}

	if ingestionFlow.Load == nil {
		statusLog.With("status", "skipped").With("reason", "unsupported").Info(fmt.Sprintf("Unsupported file types: %s", filetype))
		return nil, fmt.Errorf("%w (file %q)", &documentloader.UnsupportedFileTypeError{FileType: filetype}, opts.FileMetadata.AbsolutePath)
	}

	start := time.Now()
	checksum := sha256.Sum256(content)
	slog.Debug("File checksum calculated", "size", len(content), "duration", time.Since(start))
	opts.FileMetadata.Checksum = fmt.Sprintf("%x", checksum)

	// Mandatory Transformation: Add filename to metadata -> append extraMetadata, but do not override filename or absPath
	metadata := map[string]any{"filename": filename, "absPath": opts.FileMetadata.AbsolutePath, "fileSize": opts.FileMetadata.Size, "embeddingModel": s.EmbeddingModelProvider.EmbeddingModelName(), "fileChecksum": fmt.Sprintf("%x", checksum)}
	for k, v := range opts.ExtraMetadata {
		if _, ok := metadata[k]; !ok {
			metadata[k] = v
		}
	}

	// Reuse existing file if possible
	// TODO: this should honor textsplitter and loading settings somehow to allow for changing them and not use the existing embeddings and documents data
	var docs []vs.Document
	if opts.ReuseFiles {
		slog.Info("Checking if existing file can be reused", "checksum", opts.FileMetadata.Checksum)

		fs, err := s.Index.FindFilesByMetadata(ctx, "", types.FileMetadata{Checksum: opts.FileMetadata.Checksum}, false)
		if err != nil {
			slog.Debug("Failed to find file with checksum", "error", err, "checksum", metadata["checksum"])
		} else if len(fs) > 0 {
		fileLoop:
			for _, f := range fs {
				// check if the dataset embeddingsconfig matches - if not, we don't have to fetch the documents for this file
				ds, err := s.GetDataset(ctx, f.Dataset, nil)
				if err != nil || ds == nil {
					slog.Debug("Failed to get dataset", "error", err)
					continue
				}
				if ds.EmbeddingsProviderConfig == nil {
					slog.Debug("Dataset has no embeddings provider config")
					continue
				}
				dsEmbeddingProvider, err := embeddings.ProviderFromConfig(*ds.EmbeddingsProviderConfig)
				if err != nil {
					slog.Debug("Failed to get embeddings model provider", "error", err)
					continue
				}

				if dsEmbeddingProvider.EmbeddingModelName() != s.EmbeddingModelProvider.EmbeddingModelName() {
					slog.Debug("Embeddings model mismatch", "dataset", ds.ID, "attached", dsEmbeddingProvider.EmbeddingModelName(), "configured", s.EmbeddingModelProvider.EmbeddingModelName())
					continue
				}

				nf, err := s.FindFile(ctx, f)
				if err != nil || nf == nil {
					slog.Debug("Failed to get file", "error", err)
					continue
				}
				f = *nf
				slog.Info("Found existing file that could be reused", "fileID", f.ID, "checksum", opts.FileMetadata.Checksum, "documents", len(f.Documents), "dataset", f.Dataset)

				docs = make([]vs.Document, len(f.Documents))
				for i, existingDoc := range f.Documents {
					document, err := s.Vectorstore.GetDocument(ctx, existingDoc.ID, f.Dataset)
					if err != nil {
						// At this point we have to do all or none, as a new ingestion process could generate different chunks, e.g. due to different tokenization in a VLM processing step
						slog.Info("Failed to get document, aborting file embedding reuse process", "error", err, "docID", existingDoc.ID)
						continue fileLoop
					}
					docs[i] = vs.Document{
						ID:        uuid.NewString(),  // new UUID for the document to avoid collisions with reused docs
						Metadata:  document.Metadata, // some keys will be overridden in the transformers
						Content:   document.Content,
						Embedding: document.Embedding,
					}
				}

				slog.Info("Reused existing file", "fileID", f.ID, "checksum", opts.FileMetadata.Checksum, "documents", len(docs), "dataset", f.Dataset)
				break
			}
		}
	}

	em := &transformers.ExtraMetadata{Metadata: metadata}
	ingestionFlow.Transformations = append(ingestionFlow.Transformations, em)

	// Only run ingestion flow if we're not re-using the details of an existing file and its documents
	if len(docs) == 0 {
		docs, err = ingestionFlow.Run(ctx, bytes.NewReader(content), filename)
		if err != nil {
			statusLog.With("status", "failed").Error("Ingestion Flow failed", "error", err)
			return nil, fmt.Errorf("ingestion flow failed for file %q: %w", filename, err)
		}

		if len(docs) == 0 {
			statusLog.With("status", "skipped").Info("Ingested document", "num_documents", 0)
			return nil, nil
		}
	} else {
		// We reused documents, so we only need to run the transformers on them, e.g. to add the metadata
		docs, err = ingestionFlow.RunTransformers(ctx, docs, statusLog)
		if err != nil {
			statusLog.With("status", "failed").Error("Failed to run transformers on reused documents", "error", err)
			return nil, fmt.Errorf("failed to run transformers on reused documents: %w", err)
		}
	}

	// Sort documents
	vs.SortAndEnsureDocIndex(docs)

	// Before adding doc, we need to remove the existing documents for duplicates or old contents
	statusLog.With("component", "vectorstore").With("action", "remove").Debug("Removing existing documents")
	where := map[string]string{
		"absPath": opts.FileMetadata.AbsolutePath,
	}
	if err := s.Vectorstore.RemoveDocument(ctx, "", datasetID, where, nil); err != nil {
		statusLog.With("status", "failed").With("component", "vectorstore").Error("Failed to remove existing documents", "error", err)
		return nil, err
	}

	// Add documents to VectorStore -> This generates the embeddings
	slog.Debug("Ingesting documents", "count", len(docs), "dataset", datasetID, "file", filename)

	statusLog = statusLog.With("num_documents", len(docs))
	ctx = log.ToCtx(ctx, statusLog)

	if opts.ReuseEmbeddings {
		slog.Debug("Checking if existing embeddings can be reused", "count", len(docs))
		for i, doc := range docs {
			if len(doc.Embedding) > 0 {
				continue
			}
			existingDocs, err := s.Vectorstore.GetDocuments(ctx, "", nil, []cg.WhereDocument{
				{
					Operator: cg.WhereDocumentOperatorEquals,
					Value:    doc.Content,
				},
			})
			if err != nil {
				slog.Debug("failed to get documents for reuse", "error", err)
				continue
			}
			if len(existingDocs) == 0 {
				slog.Debug("no existing documents found for reuse")
				continue
			}
			for _, existingDoc := range existingDocs {
				if emb, ok := existingDoc.Metadata["embeddingModel"]; ok {
					if emb == s.EmbeddingModelProvider.EmbeddingModelName() {
						slog.Info("Reusing existing embedding", "docID", existingDoc.ID, "embeddingModelMeta", emb, "configuredModel", s.EmbeddingModelProvider.EmbeddingModelName())
						docs[i].Embedding = existingDoc.Embedding
						break
					} else {
						slog.Debug("not using existing embedding", "docID", existingDoc.ID, "embeddingModel", emb, "configuredModel", s.EmbeddingModelProvider.EmbeddingModelName())
						continue
					}
				}

				existingDocumentDataset, err := s.GetDatasetForDocument(ctx, existingDoc.ID)
				if err != nil {
					slog.Debug("failed to get document dataset", "error", err)
					continue
				}

				if existingDocumentDataset.EmbeddingsProviderConfig != nil {
					existingEmbeddingProvider, err := embeddings.ProviderFromConfig(*existingDocumentDataset.EmbeddingsProviderConfig)
					if err != nil {
						slog.Debug("failed to get embeddings model provider", "error", err)
						continue
					}
					if existingEmbeddingProvider.EmbeddingModelName() == s.EmbeddingModelProvider.EmbeddingModelName() {
						slog.Info("Reusing existing embedding", "docID", existingDoc.ID, "embeddingModel", existingEmbeddingProvider.EmbeddingModelName(), "configuredModel", s.EmbeddingModelProvider.EmbeddingModelName())
						docs[i].Embedding = existingDoc.Embedding
						break
					} else {
						slog.Debug("not using existing embedding", "docID", existingDoc.ID, "embeddingModel", existingEmbeddingProvider.EmbeddingModelName(), "configuredModel", s.EmbeddingModelProvider.EmbeddingModelName())
						continue
					}
				}
			}
		}
	}

	statusLog.Debug("Adding documents to vectorstore")
	startTime := time.Now()
	docIDs, err := s.Vectorstore.AddDocuments(ctx, docs, datasetID)
	if err != nil {
		statusLog.With("component", "vectorstore").With("status", "failed").With("error", err.Error()).Error("Failed to add documents")
		return nil, fmt.Errorf("failed to add documents from file %q: %w", opts.FileMetadata.AbsolutePath, err)
	}
	statusLog.Debug("Added documents to vectorstore", "duration", time.Since(startTime))

	// Record file and documents in database
	dbDocs := make([]types.Document, len(docIDs))
	for idx, docID := range docIDs {
		dbDocs[idx] = types.Document{
			ID:      docID,
			FileID:  fileID,
			Dataset: datasetID,
			Index:   idx,
		}
	}

	dbFile := types.File{
		ID:        fileID,
		Dataset:   datasetID,
		Documents: dbDocs,
		FileMetadata: types.FileMetadata{
			Name: filename,
		},
	}

	if opts.FileMetadata != nil {
		dbFile.FileMetadata.AbsolutePath = opts.FileMetadata.AbsolutePath
		dbFile.FileMetadata.Size = opts.FileMetadata.Size
		dbFile.FileMetadata.ModifiedAt = opts.FileMetadata.ModifiedAt
		dbFile.FileMetadata.Checksum = opts.FileMetadata.Checksum
	}

	iLog := statusLog.With("component", "index")
	iLog.Info("Inserting file and documents into index")
	startTime = time.Now()
	err = s.Index.CreateFile(ctx, dbFile)
	if err != nil {
		iLog.With("status", "failed").With("error", err).Error("Failed to create file in Index")
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	iLog.Info("Created file in index", "duration", time.Since(startTime))

	statusLog.With("status", "finished").Info("Ingested document", "num_documents", len(docIDs), "absolute_path", dbFile.FileMetadata.AbsolutePath, "ingestionTime", time.Since(ingestionStart))

	return docIDs, nil
}
