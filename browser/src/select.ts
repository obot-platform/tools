import { type Page } from 'playwright'
import { inspectForSelect } from './browse.ts'
import { ModelProviderCredentials } from './session.ts'

export async function select (page: Page, modelProviderCredentials: ModelProviderCredentials | undefined, model: string, userInput: string, option: string): Promise<void> {
  const selection = await inspectForSelect(page, modelProviderCredentials, model, option, userInput)

  try {
    await page.selectOption(`${selection.selector}`, selection.option, { timeout: 5000 })
  } catch (e) {
    console.error('failed to select option: ', e)
  }
}
