
import { test, expect } from './test';


test('add metadata', async ({ page, metadata }) => {
    await page.goto('/#/documents');

    for (const [key, values] of metadata) {
        await page.locator('role=menuitem >> text=metadata').click();
        await page.locator('[aria-label="Create"]').click();
        await page.fill('#key', key);
        await page.fill('#description', 'empty');

        await page.locator('button[type="submit"]').click();
        await page.waitForTimeout(100);
        expect(page.locator('text=metadata key name')).toBeVisible();

        for (const value of values) {
            await page.locator('[aria-label="Create"]').click();
            await page.waitForTimeout(100);

            await page.fill('#value', value);
            await page.click(' button[type="submit"] >> nth=1')
            await page.waitForTimeout(200);
            // check value is found in table
            expect(page.locator(`tr >> td >> text=${value}`)).toBeDefined();
        }
            // check number of values matches
            const headers = await page.locator('th').count();
            const totalData = await page.locator('td').count();
            const totalKeys = totalData / headers;
            expect(totalKeys.toString()).toMatch(values.length.toString());
    }

    // check number of keys matches 
    await page.locator('role=menuitem >> text=metadata').click();
    const headers = await page.locator('th').count();
    const totalData = await page.locator('td').count();
    const totalKeys = totalData / headers;
    expect(totalKeys.toString()).toMatch(metadata.size.toString());
});

