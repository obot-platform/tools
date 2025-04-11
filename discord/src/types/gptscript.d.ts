declare module '@gptscript-ai/gptscript' {
  export class GPTScript {
    addDatasetElements(elements: Array<{
      name: string;
      description: string;
      contents: string;
    }>, options: { name: string }): Promise<string>;
  }
} 