/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package types

// TODO - should move to common module types location - e.g. in orchestrator
type ClientDbInfo struct {
	StoreName  string // name of the mongodb collection to use for client documents
	TagMeta    string // attribute key name for the json data of a client document
	TagContent string // attribute key name for the file data of a client document
	TagContext string // attribute key name for context object in App Context
}
