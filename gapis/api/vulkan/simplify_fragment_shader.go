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
	"fmt"
	"io/ioutil"

	"github.com/google/gapid/core/log"
	"github.com/google/gapid/core/os/device"
	"github.com/google/gapid/gapis/api"
	"github.com/google/gapid/gapis/api/transform"
	"github.com/google/gapid/gapis/shadertools"
)

const opEntryPoint uint32 = 15
const opFragmentExecutionMode uint32 = 4
const constantColorShaderPath string = "/usr/local/google/home/mikaelpessa/work/gapid/gapis/shadertools/constant_color.frag"

func isFragmentShader(ctx context.Context, info VkShaderModuleCreateInfo, l *device.MemoryLayout, s *api.GlobalState) bool {
	codeSize := uint64(info.CodeSize()) / 4
	code := info.PCode().Slice(0, codeSize, l).MustRead(ctx, nil, s, nil)

	i := uint64(5) // Instructions start at word 5
	for i < codeSize {
		instruction := code[i]                 // uint32
		wordCount := uint64(instruction >> 16) // Upper 16 bits
		opCode := (instruction << 16) >> 16    // Lower 16 bits

		if opCode == opEntryPoint {
			return code[i+1] == opFragmentExecutionMode
		}

		i += wordCount
	}

	panic("No shader entry point found.")
}

func loadShader(shaderPath string) string {
	codeBytes, err := ioutil.ReadFile(shaderPath)
	if err != nil {
		fmt.Print(err)
		panic(err)
	}

	opts := shadertools.ConvertOptions{
		ShaderType:        shadertools.TypeFragment,
		MakeDebuggable:    false,
		CheckAfterChanges: true,
		Disassemble:       true,
	}

	res, err := shadertools.ConvertGlsl(string(codeBytes), &opts)
	if err != nil {
		fmt.Print(err)
		panic(err)
	}

	return res.SourceCode
}

// replaces all fragment shaders with a constant color shader
func simplifyFragmentShader(ctx context.Context) transform.Transformer {
	ctx = log.Enter(ctx, "simplifyFragmentShader")
	return transform.Transform("simplifyFragmentShader", func(ctx context.Context,
		id api.CmdID, cmd api.Cmd, out transform.Writer) {

		s := out.State()
		l := s.MemoryLayout
		cb := CommandBuilder{Thread: cmd.Thread(), Arena: s.Arena}
		switch cmd := cmd.(type) {
		case *VkCreateShaderModule:
			oldCreateInfo := cmd.PCreateInfo().MustRead(ctx, cmd, s, nil)
			isFragment := isFragmentShader(ctx, oldCreateInfo, l, s)
			if isFragment {
				shader := loadShader(constantColorShaderPath)
				log.I(ctx, shader)
				/*
					createInfo := NewVkShaderModuleCreateInfo(
						s.Arena,
						oldCreateInfo.SType(),    // sType
						oldCreateInfo.PNext(),    // pNext
						oldCreateInfo.Flags(),    // flags
						oldCreateInfo.CodeSize(), // codeSize
						oldCreateInfo.PCode(),    // pCode
					)
				*/
				createInfoData := s.AllocDataOrPanic(ctx, oldCreateInfo)
				defer createInfoData.Free()

				newCmd := cb.VkCreateShaderModule(
					cmd.Device(),
					createInfoData.Ptr(),
					cmd.PAllocator(),
					cmd.PShaderModule(),
					VkResult_VK_SUCCESS,
				).AddRead(createInfoData.Data())

				out.MutateAndWrite(ctx, id, newCmd)
			} else {
				out.MutateAndWrite(ctx, id, cmd)
			}
		default:
			out.MutateAndWrite(ctx, id, cmd)
		}
	})
}
