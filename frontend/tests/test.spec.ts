
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
            await page.fill('#value', value.name);
            if (value.automaticMatching) {
                await page.click('#match_documents');
            }
            if (value.matchBy == 'exact') {
                await page.click('#match_type_exact');
            } else {
                await page.click('#match_type_regex');
            }
            await page.fill('#match_filter', value.filter);

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


test('match metadata ids', async ({ page, metadata, metadataKeys, context }) => {
    const storage = await context.storageState();
    // @ts-ignore
    const token = storage.origins[0].localStorage.find(item => item.name === 'auth').value;
    // metadata keys
    const request = await page.context().request.get("/api/v1/metadata/keys", { headers: { 'Authorization': `Bearer ${token}` } });
    const data = await request.json();
    data.map(rawKey => {
        let item = metadataKeys.get(rawKey.key);
        if (item) {
            item.id = rawKey.id;
            metadataKeys.set(rawKey.key, item);
        }
    })

    // metadata values
    for (let [key, values] of metadata) {
        const keyId = metadataKeys.get(key);

        const request = await page.context().request.get(`/api/v1/metadata/keys/${keyId?.id}/values`,
            { headers: { 'Authorization': `Bearer ${token}` } });
        const data = await request.json();

        data.map(rawData => {

            let item = values.find(value => value.name == rawData.value);
            // @ts-ignore
            item.id = rawData.id;
        })
    }
})


test('add rule', async ({ page, metadata }) => {

    await page.goto('/#')
    await page.locator('role=menuitem >> text=processing').click();





})