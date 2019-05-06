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

// sets the size of all textures to 1x1x1
func minimizeTextureSize(ctx context.Context) transform.Transformer {
	ctx = log.Enter(ctx, "Minimize texture")
	return transform.Transform("Minimize texture", func(ctx context.Context,
		id api.CmdID, cmd api.Cmd, out transform.Writer) {

		s := out.State()
		cb := CommandBuilder{Thread: cmd.Thread(), Arena: s.Arena}
		switch cmd := cmd.(type) {
		case *VkCreateImage:
			cmd.Extras().Observations().ApplyReads(s.Memory.ApplicationPool())

			imageCreateInfo := cmd.PCreateInfo().MustRead(ctx, cmd, s, nil)
			imageCreateInfo.SetExtent(NewVkExtent3D(s.Arena /* width */, 1 /* height */, 1 /* depth */, 1))

			imageCreateInfoData := s.AllocDataOrPanic(ctx, imageCreateInfo)
			defer imageCreateInfoData.Free()

			newCmd := cb.VkCreateImage(
				cmd.Device(),
				imageCreateInfoData.Ptr(),
				cmd.PAllocator(),
				cmd.PImage(),
				VkResult_VK_SUCCESS,
			).AddRead(imageCreateInfoData.Data())

			for _, w := range cmd.Extras().Observations().Writes {
				newCmd.AddWrite(w.Range, w.ID)
			}
			out.MutateAndWrite(ctx, id, newCmd)
		default:
			out.MutateAndWrite(ctx, id, cmd)
		}
	})
}
