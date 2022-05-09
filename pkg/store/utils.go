// Copyright (c) 2022 Lukasz Janyst <lukasz@jany.st>
//
// This file is part of Peroxide.
//
// Peroxide is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Peroxide is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Peroxide.  If not, see <https://www.gnu.org/licenses/>.

package store

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ljanyst/peroxide/pkg/files"
)

const Version = "c11"

type StoreInfo struct {
	StoreVersion string
}

func recreateStore(storeDir, storeInfoPath string) error {
	if err := files.Remove(storeDir).Do(); err != nil {
		return err
	}

	if err := os.MkdirAll(storeDir, 0700); err != nil {
		return err
	}

	f, err := os.Create(storeInfoPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(StoreInfo{Version})
}

func ClearIncompatibleStore(storeDir string) error {
	storeInfoPath := filepath.Join(storeDir, "store_info.json")
	data, err := ioutil.ReadFile(storeInfoPath)
	if err != nil {
		return recreateStore(storeDir, storeInfoPath)
	}

	var info StoreInfo
	if err := json.Unmarshal(data, &info); err != nil {
		// Nuke the store if we can't unmarshal the info
		return recreateStore(storeDir, storeInfoPath)
	}

	if info.StoreVersion != Version {
		return recreateStore(storeDir, storeInfoPath)
	}

	return nil
}
