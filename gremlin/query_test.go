/*
 * Copyright (C) 2018 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy ofthe License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specificlanguage governing permissions and
 * limitations under the License.
 *
 */

package gremlin

import (
	"fmt"
	"testing"
)

func TestQueryBuilder(t *testing.T) {
	const key, val = "Type", "host"
	expected := fmt.Sprintf(`G.V().Has("%s", Regex("%s"), "%s", Regex("%s.*"), "%s", Regex(".*%s")).Flows().Sort()`,
		key, val, key, val, key, val)
	actual := G.V().Has(Quote(key), Regex(val), Quote(key), Regex("%s.*", val), Quote(key), Regex(".*%s", val)).Flows().Sort()
	if actual.String() != expected {
		t.Errorf("Wrong query,\nexpected: \"%s\",\nactual: \"%s\"", expected, actual)
	}
}
