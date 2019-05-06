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

package vulkan

import (
	"context"

	"github.com/google/gapid/core/log"
	"github.com/google/gapid/gapis/api"
	"github.com/google/gapid/gapis/api/transform"
)

// sets the number of vertices to be drawn to 1
func setVertexCountToOne(ctx context.Context) transform.Transformer {
	ctx = log.Enter(ctx, "setVertexCountToOne")
	return transform.Transform("setVertexCountToOne", func(ctx context.Context,
		id api.CmdID, cmd api.Cmd, out transform.Writer) {

		s := out.State()
		cb := CommandBuilder{Thread: cmd.Thread(), Arena: s.Arena}
		switch cmd := cmd.(type) {
		case *VkCmdDraw:
			newCmd := cb.VkCmdDraw(cmd.commandBuffer,
				/* vertex count */ 1,
				/* instance count */ 1,
				cmd.FirstVertex(),
				cmd.FirstInstance())

			out.MutateAndWrite(ctx, id, newCmd)
		default:
			out.MutateAndWrite(ctx, id, cmd)
		}
	})
}
