// Copyright 2017 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package build

import (
	"bytes"
	"fmt"
	"strings"
)

// DumpMakeVars can be used to extract the values of Make variables after the
// product configurations are loaded. This is roughly equivalent to the
// `get_build_var` bash function.
//
// goals can be used to set MAKECMDGOALS, which emulates passing arguments to
// Make without actually building them. So all the variables based on
// MAKECMDGOALS can be read.
//
// vars is the list of variables to read. The values will be put in the
// returned map.
func DumpMakeVars(ctx Context, config Config, goals, vars []string) (map[string]string, error) {
	return dumpMakeVars(ctx, config, goals, vars, false)
}

func dumpMakeVars(ctx Context, config Config, goals, vars []string, write_soong_vars bool) (map[string]string, error) {
	ctx.BeginTrace("dumpvars")
	defer ctx.EndTrace()

	cmd := Command(ctx, config, "dumpvars",
		config.PrebuiltBuildTool("ckati"),
		"-f", "build/make/core/config.mk",
		"--color_warnings",
		"--kati_stats",
		"dump-many-vars",
		"MAKECMDGOALS="+strings.Join(goals, " "))
	cmd.Environment.Set("CALLED_FROM_SETUP", "true")
	cmd.Environment.Set("BUILD_SYSTEM", "build/make/core")
	if write_soong_vars {
		cmd.Environment.Set("WRITE_SOONG_VARIABLES", "true")
	}
	cmd.Environment.Set("DUMP_MANY_VARS", strings.Join(vars, " "))
	cmd.Sandbox = dumpvarsSandbox
	output := bytes.Buffer{}
	cmd.Stdout = &output
	pipe, err := cmd.StderrPipe()
	if err != nil {
		ctx.Fatalln("Error getting output pipe for ckati:", err)
	}
	cmd.StartOrFatal()
	// TODO: error out when Stderr contains any content
	katiRewriteOutput(ctx, pipe)
	cmd.WaitOrFatal()

	ret := make(map[string]string, len(vars))
	for _, line := range strings.Split(output.String(), "\n") {
		if len(line) == 0 {
			continue
		}

		if key, value, ok := decodeKeyValue(line); ok {
			if value, ok = singleUnquote(value); ok {
				ret[key] = value
				ctx.Verboseln(key, value)
			} else {
				return nil, fmt.Errorf("Failed to parse make line: %q", line)
			}
		} else {
			return nil, fmt.Errorf("Failed to parse make line: %q", line)
		}
	}

	return ret, nil
}

// Variables to print out in the top banner
var BannerVars = []string{
	"PLATFORM_VERSION_CODENAME",
	"PLATFORM_VERSION",
	"CANDY_VERSION",
	"TARGET_PRODUCT",
	"TARGET_BUILD_VARIANT",
	"TARGET_BUILD_TYPE",
	"TARGET_BUILD_APPS",
	"TARGET_ARCH",
	"TARGET_ARCH_VARIANT",
	"TARGET_CPU_VARIANT",
	"TARGET_2ND_ARCH",
	"TARGET_2ND_ARCH_VARIANT",
	"TARGET_2ND_CPU_VARIANT",
	"HOST_ARCH",
	"HOST_2ND_ARCH",
	"HOST_OS",
	"HOST_OS_EXTRA",
	"HOST_CROSS_OS",
	"HOST_CROSS_ARCH",
	"HOST_CROSS_2ND_ARCH",
	"HOST_BUILD_TYPE",
	"BUILD_ID",
	"OUT_DIR",
	"AUX_OS_VARIANT_LIST",
	"TARGET_BUILD_PDK",
	"PDK_FUSION_PLATFORM_ZIP",
	"PRODUCT_SOONG_NAMESPACES",
	"TARGET_USE_SDCLANG",
	"RECOVERY_VARIANT",
	"WITH_SU",
	"WITH_GMS",
	"TARGET_GCC_VERSION",
}

func Banner(make_vars map[string]string) string {
	b := &bytes.Buffer{}

    fmt.Fprintln(b, "               'Wb.              .dW'               ")
    fmt.Fprintln(b, "                .dWK0XXK00000KXO0Wb.                ")
    fmt.Fprintln(b, "               WKko''''.......''''dOXW.             ")
    fmt.Fprintln(b, "            NOl;.  .,.        .,.   .;oON           ")
    fmt.Fprintln(b, "         W0l'      .'.         ..       ,oKW        ")
    fmt.Fprintln(b, "       WO:.                               .lK       ")
    fmt.Fprintln(b, "     .0l.           .';::::;,..             'xN     ")
    fmt.Fprintln(b, "    dK'         .;ok00OOkkkOO00xc.           .lX.   ")
    fmt.Fprintln(b, "   KN;        ,d00xc,..     ..;dKKo.           lN   ")
    fmt.Fprintln(b, " .Wl        .dX0l.              .dXO,           dW  ")
    fmt.Fprintln(b, " lX.       'OXd.      .,coool;.   lNO.          .oK ")
    fmt.Fprintln(b, " Xo       .kNo.      ;OXNWMMMWk'  'ON:           xN.")
    fmt.Fprintln(b, ".Wl       cNO.      '0XxOWMMMMWl  '0Nc           XN ")
    fmt.Fprintln(b, ".WD       oWx.      ,KXllKNNWXd. .oNO.           XW.")
    fmt.Fprintln(b, " do       oWx.       lX0l;;;;'.'cOXx'           .xN ")
    fmt.Fprintln(b, " lX'      ;XK,        ,d000OkkOK0d,             cX' ")
    fmt.Fprintln(b, "  WO      .xNd.         .';;::;'.              ;KW  ")
    fmt.Fprintln(b, "  'XC      'OXo.                              :Kl   ")
    fmt.Fprintln(b, "   'XW      .kXx'                           .oXN    ")
    fmt.Fprintln(b, "      WX,    .lKKo'                       .l0W      ")
    fmt.Fprintln(b, "        Xx,    .o0Kx:.                 .:xKW        ")
    fmt.Fprintln(b, "          NOc'   .;d00Odl:;,'...'';:ldOXW           ")
    fmt.Fprintln(b, "            WXOo;.. .':oxkOOOO000KNWNK'             ")
    fmt.Fprintln(b, "                dWKxc,.    ..  cox0NW               ")
    fmt.Fprintln(b, "                     dXKOBN0K0M[k:'                 ")
    fmt.Fprintln(b, "                                                    ")
    fmt.Fprintln(b, "           CANDYROMS - TASTE THE SWEETNESS!         ")
    fmt.Fprintln(b, "=============================================================")
    fmt.Fprintf(b, "%s=%s\n", "HOST_ARCH", make_vars["HOST_ARCH"])
    fmt.Fprintf(b, "%s=%s\n", "HOST_2ND_ARCH", make_vars["HOST_2ND_ARCH"])
    fmt.Fprintf(b, "%s=%s\n", "HOST_OS", make_vars["HOST_OS"])
    fmt.Fprintf(b, "%s=%s\n", "HOST_OS_EXTRA", make_vars["HOST_OS_EXTRA"])
    fmt.Fprintf(b, "%s=%s\n", "HOST_CROSS_OS", make_vars["HOST_CROSS_OS"])
    fmt.Fprintf(b, "%s=%s\n", "HOST_CROSS_ARCH", make_vars["HOST_CROSS_ARCH"])
    fmt.Fprintf(b, "%s=%s\n", "HOST_CROSS_2ND_ARCH", make_vars["HOST_CROSS_2ND_ARCH"])
    fmt.Fprintln(b, "=============================================================")
    fmt.Fprintf(b, "%s=%s\n", "CANDY_VERSION", make_vars["CANDY_VERSION"])
    fmt.Fprintf(b, "%s=%s\n", "BUILD_ID", make_vars["BUILD_ID"])
    fmt.Fprintf(b, "%s=%s\n", "OUT_DIR", make_vars["OUT_DIR"])
    fmt.Fprintln(b, "=============================================================")
    fmt.Fprintf(b, "%s=%s\n", "PLATFORM_VERSION_CODENAME", make_vars["PLATFORM_VERSION_CODENAME"])
    fmt.Fprintf(b, "%s=%s\n", "PLATFORM_VERSION", make_vars["PLATFORM_VERSION"])
    fmt.Fprintf(b, "%s=%s\n", "TARGET_PRODUCT", make_vars["TARGET_PRODUCT"])
    fmt.Fprintf(b, "%s=%s\n", "TARGET_BUILD_VARIANT", make_vars["TARGET_BUILD_VARIANT"])
    fmt.Fprintf(b, "%s=%s\n", "TARGET_BUILD_TYPE", make_vars["TARGET_BUILD_TYPE"])
    fmt.Fprintf(b, "%s=%s\n", "TARGET_GCC_VERSION", make_vars["TARGET_GCC_VERSION"])
    fmt.Fprintf(b, "%s=%s\n", "TARGET_ARCH", make_vars["TARGET_ARCH"])
    fmt.Fprintf(b, "%s=%s\n", "TARGET_ARCH_VARIANT", make_vars["TARGET_ARCH_VARIANT"])
    fmt.Fprintf(b, "%s=%s\n", "TARGET_CPU_VARIANT", make_vars["TARGET_CPU_VARIANT"])
    fmt.Fprintf(b, "%s=%s\n", "TARGET_2ND_ARCH", make_vars["TARGET_2ND_ARCH"])
    fmt.Fprintf(b, "%s=%s\n", "TARGET_2ND_ARCH_VARIANT", make_vars["TARGET_2ND_ARCH_VARIANT"])
    fmt.Fprintf(b, "%s=%s\n", "TARGET_2ND_CPU_VARIANT", make_vars["TARGET_2ND_CPU_VARIANT"])
    fmt.Fprintln(b, "=============================================================")

	return b.String()
}

func runMakeProductConfig(ctx Context, config Config) {
	// Variables to export into the environment of Kati/Ninja
	exportEnvVars := []string{
		// So that we can use the correct TARGET_PRODUCT if it's been
		// modified by PRODUCT-*/APP-* arguments
		"TARGET_PRODUCT",
		"TARGET_BUILD_VARIANT",
		"TARGET_BUILD_APPS",

		// compiler wrappers set up by make
		"CC_WRAPPER",
		"CXX_WRAPPER",
		"JAVAC_WRAPPER",

		// ccache settings
		"CCACHE_COMPILERCHECK",
		"CCACHE_SLOPPINESS",
		"CCACHE_BASEDIR",
		"CCACHE_CPP2",
	}

	allVars := append(append([]string{
		// Used to execute Kati and Ninja
		"NINJA_GOALS",
		"KATI_GOALS",

		// To find target/product/<DEVICE>
		"TARGET_DEVICE",
	}, exportEnvVars...), BannerVars...)

	make_vars, err := dumpMakeVars(ctx, config, config.Arguments(), allVars, true)
	if err != nil {
		ctx.Fatalln("Error dumping make vars:", err)
	}

	// Print the banner like make does
	fmt.Fprintln(ctx.Stdout(), Banner(make_vars))

	// Populate the environment
	env := config.Environment()
	for _, name := range exportEnvVars {
		if make_vars[name] == "" {
			env.Unset(name)
		} else {
			env.Set(name, make_vars[name])
		}
	}

	config.SetKatiArgs(strings.Fields(make_vars["KATI_GOALS"]))
	config.SetNinjaArgs(strings.Fields(make_vars["NINJA_GOALS"]))
	config.SetTargetDevice(make_vars["TARGET_DEVICE"])
}
