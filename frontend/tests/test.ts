
import { test as base } from '@playwright/test';


type MetadataFixture = {
    metadata: Map<string, string[]>;
}


export const test = base.extend<MetadataFixture>({
    metadata: new Map([
        ['author', ['none', 'another']],
        ['class', ['publication', 'paper', 'report']],
        ['project', ['personal', 'work', 'study']],
    ])
})

export { expect } from '@playwright/test';