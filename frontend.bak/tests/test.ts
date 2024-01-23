
import { test as base } from '@playwright/test';


type IdName = {
    id: string;
    name: string;
}


type MetadataValue = {
    id: string;
    name: string;
    automaticMatching: boolean;
    matchBy: string;
    filter: string;
}

type Rule = {
    id: string;
    name: string;
    description: string;
    mode: string;
    conditions: RuleCondition[];
    actions: RuleAction[];
}

type RuleCondition = {
    enabled: boolean;
    caseInsensitive: boolean;
    inverted: boolean;
    conditionType: string;
    isRegex: boolean;
    value: string;
    dateFmt: string;
    metadata: any;
}

type RuleAction = {
    enabled: boolean;
    onCondition: boolean;
    action: string;
    value: string;
    metadata: any;
}


type MetadataFixture = {
    // set IDs when creating metadata
    metadataKeys: Map<string, IdName>;

    metadata: Map<string, MetadataValue[]>;
    rules: Rule[];
}


export let test = base.extend<MetadataFixture>({
    metadataKeys: new Map([
        ['author', { id: '0', name: 'author' }],
        ['class', { id: '0', name: 'class' }],
        ['project', { id: '0', name: 'project' }],

    ]),
    metadata: new Map([
        ['author', [
            {
                id: '0',
                name: 'unknown',
                automaticMatching: false,
                matchBy: "exact",
                filter: "",
            },
            {
                id: '0',
                name: 'another',
                automaticMatching: true,
                matchBy: "regex",
                filter: "another author",
            },
            {
                id: '0',
                name: 'lorem',
                automaticMatching: false,
                matchBy: "exact",
                filter: "",
            }
        ]],
        ['class', [
            {
                id: '0',
                name: 'publication',
                automaticMatching: false,
                matchBy: "exact",
                filter: "",
            },
            {
                id: '0',
                name: 'paper',
                automaticMatching: false,
                matchBy: "exact",
                filter: "",

            },
            {
                id: '0',
                name: 'report',
                automaticMatching: false,
                matchBy: "exact",
                filter: "",
            }
        ]],
        ['project', [
            {
                id: '0',
                name: 'personal',
                automaticMatching: false,
                matchBy: "exact",
                filter: "",
            }, {
                id: '0',
                name: 'work',
                automaticMatching: false,
                matchBy: "exact",
                filter: "",


            }, {
                id: '0',
                name: 'study',
                automaticMatching: false,
                matchBy: "exact",
                filter: "",
            }]],
    ]),
    rules: [
        {
            id: '0',
            name: "no author",
            description: "",
            mode: "match all",
            conditions: [
                {
                    enabled: true,
                    caseInsensitive: false,
                    inverted: false,
                    conditionType: "metadata_has_key",
                    isRegex: false,
                    value: "",
                    dateFmt: "",
                    metadata: {
                        key_id: 0,
                        value: 1,
                    }
                },
            ],
            actions: [
                {
                    enabled: true,
                    onCondition: true,
                    action: 'metadata_add',
                    value: "",
                    metadata: {
                        key_id: 0,
                        value: 1,
                    }
                }
            ],
        }
    ]

})

export { expect } from '@playwright/test';