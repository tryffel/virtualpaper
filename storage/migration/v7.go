/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2021  Tero Vierimaa
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


const schemaV7 = `
CREATE TABLE rules (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL,
	name TEXT NOT NULL DEFAULT '',
	description TEXT NOT NULL DEFAULT '',
	enabled BOOLEAN NOT NULL DEFAULT true,
	rule_order INT NOT NULL,
	mode INT NOT NULL,
	created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ,

	CONSTRAINT fk_user FOREIGN KEY(user_id)
        REFERENCES users(id),
	CONSTRAINT rules_user_order_unique UNIQUE (user_id, rule_order)
);


CREATE TABLE rule_conditions (
	id SERIAL PRIMARY KEY,
	rule_id INT NOT NULL,
	enabled BOOLEAN NOT NULL DEFAULT true,
	case_insensitive BOOLEAN NOT NULL DEFAULT true,
	inverted_match BOOLEAN NOT NULL DEFAULT false,
	condition_type TEXT NOT NULL DEFAULT '',
	is_regex BOOL NOT NULL DEFAULT false,
	value TEXT NOT NULL DEFAULT '',
	metadata_key INT,
	metadata_value INT,

	CONSTRAINT fk_rule FOREIGN KEY(rule_id)
        REFERENCES rules(id),
	CONSTRAINT fk_metadata_key FOREIGN KEY(metadata_key)
        REFERENCES metadata_keys(id),
	CONSTRAINT fk_metadata_value FOREIGN KEY(metadata_value)
        REFERENCES metadata_values(id)
);


CREATE TABLE rule_actions (
	id SERIAL PRIMARY KEY,
	rule_id INT NOT NULL,
	enabled BOOLEAN NOT NULL DEFAULT true,
	on_condition BOOLEAN NOT NULL DEFAULT true,
	action TEXT NOT NULL DEFAULT '',
	value TEXT NOT NULL DEFAULT '',
	metadata_key INT,
	metadata_value INT,

	CONSTRAINT fk_rule FOREIGN KEY(rule_id)
        REFERENCES rules(id),
	CONSTRAINT fk_metadata_key FOREIGN KEY(metadata_key)
        REFERENCES metadata_keys(id),
	CONSTRAINT fk_metadata_value FOREIGN KEY(metadata_value)
        REFERENCES metadata_values(id)
);
`
