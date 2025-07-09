from typing import List, Tuple, Dict, Any
from io import BytesIO
from markitdown import MarkItDown, StreamInfo, DocumentConverterResult


def extract_text_from_content(
    content: bytes, file_extension: str
) -> DocumentConverterResult:
    """Extract text from file content based on content type use Markitdown."""
    md = MarkItDown(enable_plugins=False)

    return md.convert(
        BytesIO(content), stream_info=StreamInfo(extension=file_extension)
    )


def chunk_text(
    text: str,
    strategy: str = "sentence",
    chunk_size: int = 1000,
    chunk_overlap: int = 200,
    **kwargs,
) -> List[Tuple[str, int]]:
    """
    Split text into chunks using intelligent chunking strategies. Prefer Chonkie if available.

    Args:
        text: The text to chunk
        strategy: Chunking strategy - "sentence", "semantic", "recursive", "token", or "basic"
        chunk_size: Target chunk size (tokens for token-based, characters for others)
        chunk_overlap: Overlap between chunks
        **kwargs: Additional arguments for specific chunkers

    Returns:
        List of tuples (chunk_text, start_offset)
    """
    if not text:
        return []

    # Try to use Chonkie for intelligent chunking
    try:
        return _chunk_with_chonkie(text, strategy, chunk_size, chunk_overlap, **kwargs)
    except ImportError:
        # Fallback to improved basic chunking if Chonkie is not available
        print("Warning: Chonkie not available, falling back to improved basic chunking")
        return _chunk_basic_improved(text, chunk_size, chunk_overlap)
    except Exception as e:
        print(f"Warning: Chonkie chunking failed ({e}), falling back to basic chunking")
        return _chunk_basic_improved(text, chunk_size, chunk_overlap)


def _chunk_with_chonkie(
    text: str, strategy: str, chunk_size: int, chunk_overlap: int, **kwargs
) -> List[Tuple[str, int]]:
    """Use Chonkie library for intelligent chunking."""

    if strategy == "sentence":
        from chonkie import SentenceChunker

        chunker = SentenceChunker(chunk_size=chunk_size, chunk_overlap=chunk_overlap)
    elif strategy == "semantic":
        from chonkie import SemanticChunker

        # Semantic chunker uses embeddings to determine chunk boundaries
        chunker = SemanticChunker(
            chunk_size=chunk_size,
            similarity_threshold=kwargs.get("similarity_threshold", 0.5),
        )
    elif strategy == "recursive":
        from chonkie import RecursiveChunker

        chunker = RecursiveChunker(chunk_size=chunk_size, chunk_overlap=chunk_overlap)
    elif strategy == "token":
        from chonkie import TokenChunker

        chunker = TokenChunker(chunk_size=chunk_size, chunk_overlap=chunk_overlap)
    else:
        # Default to sentence chunker
        from chonkie import SentenceChunker

        chunker = SentenceChunker(chunk_size=chunk_size, chunk_overlap=chunk_overlap)

    # Get chunks from Chonkie
    chunks = chunker(text)

    # Convert to our format with start offsets
    result = []
    current_offset = 0

    for chunk in chunks:
        chunk_text = chunk.text
        # Find the actual start position in the original text
        start_pos = text.find(chunk_text, current_offset)
        if start_pos == -1:
            # If exact match not found, use current offset
            start_pos = current_offset

        result.append((chunk_text, start_pos))
        current_offset = start_pos + len(chunk_text)

    return result


def _chunk_basic_improved(
    text: str, chunk_size: int, chunk_overlap: int
) -> List[Tuple[str, int]]:
    """
    Improved basic chunking that respects sentence and word boundaries.
    Used as fallback when Chonkie is not available.
    """
    chunks = []
    start = 0

    while start < len(text):
        # Calculate end position
        end = start + chunk_size

        # If this isn't the last chunk, try to break at a sentence or word boundary
        if end < len(text):
            # Look for sentence endings within the last 20% of the chunk
            boundary_search_start = start + int(chunk_size * 0.8)

            # Try to find sentence boundary first
            sentence_patterns = [". ", "! ", "? ", ".\n", "!\n", "?\n"]
            best_boundary = -1

            for pattern in sentence_patterns:
                boundary = text.find(pattern, boundary_search_start, end)
                if boundary > best_boundary:
                    best_boundary = boundary + len(pattern)

            if best_boundary > boundary_search_start:
                end = best_boundary
            else:
                # If no sentence ending, look for paragraph break
                para_break = text.find("\n\n", boundary_search_start, end)
                if para_break > boundary_search_start:
                    end = para_break + 2
                else:
                    # If no paragraph break, look for line break
                    line_break = text.rfind("\n", boundary_search_start, end)
                    if line_break > boundary_search_start:
                        end = line_break + 1
                    else:
                        # Finally, try word boundary
                        word_boundary = text.rfind(" ", boundary_search_start, end)
                        if word_boundary > boundary_search_start:
                            end = word_boundary + 1

        # Extract chunk
        chunk = text[start:end].strip()
        if chunk:
            chunks.append((chunk, start))

        # Move start position (with overlap)
        if end >= len(text):
            break

        # Calculate next start with overlap
        overlap_start = max(start + 1, end - chunk_overlap)

        # Try to start overlap at a word boundary
        if overlap_start < end:
            word_start = text.find(" ", overlap_start)
            if word_start != -1 and word_start < end:
                overlap_start = word_start + 1

        start = overlap_start

        # Prevent infinite loop
        if start >= len(text):
            break

    return chunks


def get_chunking_info() -> Dict[str, Any]:
    """Get information about available chunking strategies."""
    info = {
        "strategies": {
            "sentence": "Respects sentence boundaries for natural breaks",
            "semantic": "Uses embeddings to group semantically similar content (requires Chonkie)",
            "recursive": "Tries multiple separators hierarchically (requires Chonkie)",
            "token": "Token-aware chunking with proper tokenizer support (requires Chonkie)",
            "basic": "Simple character-based chunking with improved boundary detection",
        },
        "chonkie_available": False,
    }

    try:
        import chonkie

        info["chonkie_available"] = True
        info["chonkie_version"] = getattr(chonkie, "__version__", "unknown")
    except ImportError:
        pass

    return info
