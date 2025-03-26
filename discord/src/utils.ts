import { GPTScript } from '@gptscript-ai/gptscript';

/**
 * Formats a timestamp in a human-readable format (hh:mm AM/PM Month Day, YYYY (Timezone))
 * @param timestamp The timestamp in milliseconds since epoch
 * @returns The formatted date string
 */
export function formatTime(timestamp: number | null | undefined): string | null {
  if (timestamp == null) return null;

  const tz = process.env.OBOT_USER_TIMEZONE || "UTC";
  const date = new Date(timestamp);

  try {
    const formatted = date.toLocaleString('en-US', {
      timeZone: tz,
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
      month: 'long',
      day: 'numeric',
      year: 'numeric'
    });
    return `${formatted} (${tz})`;
  } catch (e) {
    // If timezone is invalid, fall back to UTC
    const formatted = new Date(timestamp).toLocaleString('en-US', {
      timeZone: 'UTC',
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
      month: 'long',
      day: 'numeric',
      year: 'numeric'
    });
    return `${formatted} (UTC)`;
  }
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