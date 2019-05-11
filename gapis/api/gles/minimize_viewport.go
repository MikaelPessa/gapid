// Copyright (C) 2019 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gles

import (
	"context"

	"github.com/google/gapid/core/log"
	"github.com/google/gapid/gapis/api"
	"github.com/google/gapid/gapis/api/transform"
)

// minimizeViewport returns a transform that sets all viewport sizes to a 1x1 square.
func minimizeViewport(ctx context.Context) transform.Transformer {
	ctx = log.Enter(ctx, "Minimize viewport")

	const width = 1
	const height = 1

	// Per-instance variable.
	transformApplied := false

	return transform.Transform("Minimize viewport", func(ctx context.Context,
		id api.CmdID, cmd api.Cmd, out transform.Writer) {

		s := out.State()
		cmd.Extras().Observations().ApplyReads(s.Memory.ApplicationPool())
		cb := CommandBuilder{Thread: cmd.Thread(), Arena: s.Arena}
		switch cmd := cmd.(type) {
		case *GlViewport:
			out.MutateAndWrite(ctx, id, cb.GlViewport(cmd.X(), cmd.Y(), width, height))
			transformApplied = true
			return
		// Context should be properly bound by the first draw call
		case *GlDrawArrays, *GlDrawArraysIndirect, *GlDrawArraysInstanced,
			*GlDrawBuffers, *GlDrawElements, *GlDrawElementsBaseVertex,
			*GlDrawElementsIndirect, *GlDrawElementsInstanced,
			*GlDrawElementsInstancedBaseVertex, *GlDrawRangeElements,
			*GlDrawRangeElementsBaseVertex, *EglSwapBuffers:

			// Apply once outside GlViewport, in case we missed the GlViewport event.
			if !transformApplied {
				out.MutateAndWrite(ctx, id, cb.GlViewport(0, 0, width, height))
				transformApplied = true
			}
		}

		out.MutateAndWrite(ctx, id, cmd)
	})
}
