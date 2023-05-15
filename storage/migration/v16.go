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

const schemaV16 = `
ALTER TABLE process_queue 
    ADD COLUMN action_order INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN runner_id TEXT;
UPDATE process_queue SET action_order=step;

ALTER TABLE process_queue DROP COLUMN step;
ALTER TABLE process_queue ADD COLUMN action TEXT;

UPDATE process_queue SET action='hash' WHERE action_order=1;
UPDATE process_queue SET action='thumbnail' WHERE action_order=2;
UPDATE process_queue SET action='extract' WHERE action_order=3;
UPDATE process_queue SET action='rules' WHERE action_order=4;
UPDATE process_queue SET action='index' WHERE action_order=5;

ALTER TABLE process_queue ALTER COLUMN action SET NOT NULL;
`
