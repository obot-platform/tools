import { GPTScript } from '@gptscript-ai/gptscript';

/**
 * Converts a timestamp to RFC 3339 format
 * @param timestamp The timestamp in milliseconds since epoch
 * @returns The RFC 3339 formatted date string
 */
export function toRFC3339(timestamp: number | null | undefined): string | null {
  if (timestamp == null) return null;
  return new Date(timestamp).toISOString();
}

/**
 * Creates a dataset from an array of elements
 * @param elements Array of elements to add to the dataset
 * @param datasetName Base name of the dataset
 */
export async function createDataset(elements: any[], datasetName: string): Promise<void> {
  const gptscriptClient = new GPTScript();
  const datasetElements = elements.map(element => ({
    name: element.name || element.id,
    description: element.description || "",
    contents: JSON.stringify(element),
  }));

  const timestamp = new Date().getTime();
  const uniqueDatasetName = `${datasetName}_${timestamp}`;

  const datasetID = await gptscriptClient.addDatasetElements(datasetElements, {
    name: uniqueDatasetName,
  });

  console.log(`Created dataset with ID ${datasetID} with ${elements.length} elements`);
} 