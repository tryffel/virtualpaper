import { chromium, FullConfig, test, expect } from '@playwright/test';
import { start } from 'repl';


async function globalSetup(config: FullConfig) {
    const { baseURL } = config.projects[0].use;
    const browser = await chromium.launch()
    const page = await browser.newPage();
    await page.goto(`${baseURL}/#/login`);
    await page.fill('#username', 'user');
    await page.fill('#password', 'user');
    await page.click('button[type="submit"]')
    // token generation takes some time, ensure token is saved to local storage first.
    await page.waitForTimeout(3000);

    await page.context().storageState({path: 'storageState.json'});
    await browser.close();
}

export default globalSetup;