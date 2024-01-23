// playwright.config.ts
import { type PlaywrightTestConfig, devices } from '@playwright/test';

const config: PlaywrightTestConfig = {
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  globalSetup: require.resolve('./tests/global-setup'),
  use: {
    trace: 'on-first-retry',
    // Tell all tests to load signed-in state from 'storageState.json'.
    storageState: 'storageState.json',
    baseURL: 'http://localhost:8000/',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    /*{
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },*/
  ],
  
};
export default config;
