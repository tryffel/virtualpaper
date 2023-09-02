/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2023  Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package migration

const schemaV17 = `

CREATE TABLE languages (
    id TEXT PRIMARY KEY,
    name_eng TEXT NOT NULL UNIQUE
);

COMMENT ON COLUMN languages.id IS 'ISO 639-2T code';

ALTER TABLE documents
    ADD COLUMN lang TEXT REFERENCES languages(id);


-- supported languages for github.com/pemistahl/lingua-go
INSERT INTO languages (id, name_eng) VALUES 
	('af', 'Afrikaans'),
	('sq', 'Albanian'),
	('ar', 'Arabic'),
	('hy', 'Armenian'),
	('az', 'Azerbaijani'),
	('eu', 'Basque'),
	('be', 'Belarusian'),
	('bn', 'Bengali'),
	('nb', 'Bokmal'),
	('bs', 'Bosnian'),
	('bg', 'Bulgarian'),
	('ca', 'Catalan'),
	('zh', 'Chinese'),
	('hr', 'Croatian'),
	('cs', 'Czech'),
	('da', 'Danish'),
	('nl', 'Dutch'),
	('en', 'English'),
	('eo', 'Esperanto'),
	('et', 'Estonian'),
	('fi', 'Finnish'),
	('fr', 'French'),
	('lg', 'Ganda'),
	('ka', 'Georgian'),
	('de', 'German'),
	('el', 'Greek'),
	('gu', 'Gujarati'),
	('he', 'Hebrew'),
	('hi', 'Hindi'),
	('hu', 'Hungarian'),
	('is', 'Icelandic'),
	('id', 'Indonesian'),
	('ga', 'Irish'),
	('it', 'Italian'),
	('ja', 'Japanese'),
	('kk', 'Kazakh'),
	('ko', 'Korean'),
	('la', 'Latin'),
	('lv', 'Latvian'),
	('lt', 'Lithuanian'),
	('mk', 'Macedonian'),
	('ms', 'Malay'),
	('mi', 'Maori'),
	('mr', 'Marathi'),
	('mn', 'Mongolian'),
	('nn', 'Nynorsk'),
	('fa', 'Persian'),
	('pl', 'Polish'),
	('pt', 'Portuguese'),
	('pa', 'Punjabi'),
	('rm', 'Romanian'),
	('ru', 'Russian'),
	('sr', 'Serbian'),
	('sn', 'Shona'),
	('sk', 'Slovak'),
	('sl', 'Slovene'),
	('so', 'Somali'),
	('st', 'Sotho'),
	('es', 'Spanish'),
	('sw', 'Swahili'),
	('sv', 'Swedish'),
	('tl', 'Tagalog'),
	('ta', 'Tamil'),
	('te', 'Telugu'),
	('th', 'Thai'),
	('ts', 'Tsonga'),
	('tn', 'Tswana'),
	('tr', 'Turkish'),
	('uk', 'Ukrainian'),
	('ur', 'Urdu'),
	('vi', 'Vietnamese'),
	('cy', 'Welsh'),
	('xh', 'Xhosa'),
	('yo', 'Yoruba'),
	('zu', 'Zulu');
`
