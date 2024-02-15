/*
 *  Copyright 2024 Markus W Mahlberg
 *
 *  thresholds.go is part of argononefan
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/alecthomas/kong"
	"golang.org/x/exp/maps"
)

type thresholds struct {
	sync.RWMutex
	thresholds map[float32]int
	idx        []float32
}

func (t *thresholds) UnmarshalText(text []byte) error {
	t.Lock()
	defer t.Unlock()
	t.thresholds = make(map[float32]int)
	for _, val := range strings.Split(string(text), ";") {
		kv := strings.Split(val, "=")
		if len(kv) != 2 {
			return fmt.Errorf("not a key/value pair: %s", val)
		}
		f, err := strconv.ParseFloat(kv[0], 32)
		if err != nil {
			return fmt.Errorf("parsing key %s: %s", kv[0], err)
		}
		i, err := strconv.Atoi(kv[1])
		if err != nil {
			return fmt.Errorf("parsing value %s: %s", kv[1], err)
		}
		t.thresholds[float32(f)] = i
	}
	return nil
}

func (t *thresholds) AfterApply(ctx *kong.Context) error {
	t.GenerateIndex()
	return nil
}

func (t *thresholds) GetSpeed(temperature float32) int {
	t.RLock()
	defer t.RUnlock()
	for _, th := range t.idx {
		if temperature >= th {
			return t.thresholds[th]
		}
	}
	return 0
}

func (t *thresholds) GetSpeedWithHysteresis(temperature float32, hysteresis float32) int {
	t.RLock()
	defer t.RUnlock()
	for _, th := range t.idx {
		if temperature > th-hysteresis {
			return t.thresholds[th]
		}
	}
	return 0
}

func (t *thresholds) GetThreshold(temperature float32) float32 {
	t.RLock()
	defer t.RUnlock()
	for _, th := range t.idx {
		if temperature >= th {
			return th
		}
	}
	return 0
}

func (t *thresholds) GenerateIndex() {
	t.RLock()
	defer t.RUnlock()
	if t.idx == nil || len(t.idx) == 0 || len(t.idx) != len(t.thresholds) {
		t.idx = make([]float32, len(t.thresholds))
	}
	t.idx = maps.Keys(t.thresholds)
	slices.Sort(t.idx)
	slices.Reverse(t.idx)
	return
}
