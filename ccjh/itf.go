/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ccjh

import (
	"context"
	"time"
)

type JobMeta interface {
	SetResult(any)
	SetError(error)
	SetStarted(time.Time)
	SetCompleted(time.Time)
	IsCanceled() bool
}

type TargetFunc func(context.Context, map[string]any) (any, error)
