#!/usr/bin/env bash

# Get a list of apps to build by looking at the directories in cmd.
apps=("data-node")

# Set a list of all targets.
alltargets=( \
	"linux/386" "linux/amd64" "linux/arm64" \
	"linux/mips" "linux/mipsle" "linux/mips64" "linux/mips64le" \
	"darwin/amd64" \
	"windows/386" "windows/amd64"
)

help() {
	col=""
	nocol=""
	if test -t 1 ; then
		col="\033[36m"
		nocol="\033[0m"
	fi
	echo "Command line arguments:"
	echo
	echo "  -a action  Take action:"

	grep ')[ ]##' "$0" | awk 'BEGIN {FS = "\)[ ]## "}; {printf "'"$col"'%-20s'"$nocol"' %s\n", $1, $2}'

	echo "  -d         Build debug binaries"
	echo "  -T         Build all available GOOS+GOARCH combinations"
	echo "  -t list    Build specified GOOS+GOARCH combinations (comma-separated list)"

	for target in "${alltargets[@]}" ; do
		echo -e "\t$col$target$nocol"
	done

	echo "  -s suffix  Add arbitrary suffix to compiled binary names"
	echo "  -h         Show this help"
	echo
	echo "Apps to be built:"
	for app in "${apps[@]}" ; do
		echo -e "\t$col$app$nocol"
	done
	echo
	echo "Available targets:"
}

check_golang_version() {
	local goprog
	goprog="$(command -v go)"
	if test -z "$goprog" ; then
		echo "Could not find go"
		return 1
	fi

	goversion="$("$goprog" version)"
	if ! echo "$goversion" | grep -q 'go1\.18\.' ; then
		echo "Please use Go 1.18"
		echo "Using: $goprog"
		echo "Version: $goversion"
		return 1
	fi
}

parse_args() {
	# set defaults
	action=""
	gcflags=""
	dbgsuffix=""
	suffix=""
	targets=()

	while getopts 'a:ds:Tt:h' flag
	do
		case "$flag" in
		a)
			action="$OPTARG"
			;;
		d)
			gcflags="all=-N -l"
			dbgsuffix="-dbg"
			;;
		s)
			suffix="$OPTARG"
			;;
		t)
			targets=("$OPTARG")
			;;
		T)
			targets=("${alltargets[@]}")
			;;
		h)
			help
			exit 0
			;;
		*)
			echo "Invalid option: $flag"
			exit 1
			;;
		esac
	done
}

can_build() {
	local canbuild target
	canbuild=0
	target="$1" ; shift
	for compiler in "$@" ; do
		if test -z "$compiler" ; then
			continue
		fi

		if ! command -v "$compiler" 1>/dev/null ; then
			echo "$target: Cannot build. Need $compiler"
			canbuild=1
		fi
	done
	return "$canbuild"
}

set_go_flags() {
	local target
	target="$1" ; shift
	cc=""
	cgo_ldflags=""
	cgo_cxxflags=""
	cxx=""
	goarm=""
	if test "$target" == default ; then
		goarch=""
		goos=""
		osarchsuffix=""
	else
		goarch="$(echo "$target" | cut -f2 -d/)"
		goos="$(echo "$target" | cut -f1 -d/)"
		osarchsuffix="-$goos-$goarch"
	fi
	typesuffix=""
	skip=no
	if test "$action" == build ; then
		case "$target" in
		default)
			:
			;;
		darwin/*)
			cc=o64-clang
			cxx=o64-clang++
			;;
		linux/386)
			:
			;;
		linux/amd64)
			:
			;;
		linux/arm64)
			cc=aarch64-linux-gnu-gcc-9
			cxx=aarch64-linux-gnu-g++-9
			;;
		linux/mips)
			cc=mips-linux-gnu-gcc-9
			cxx=mips-linux-gnu-g++-9
			;;
		linux/mipsle)
			cc=mipsel-linux-gnu-gcc-9
			cxx=mipsel-linux-gnu-g++-9
			;;
		linux/mips64)
			cc=mips64-linux-gnuabi64-gcc-9
			cxx=mips64-linux-gnuabi64-g++-9
			;;
		linux/mips64le)
			cc=mips64el-linux-gnuabi64-gcc-9
			cxx=mips64el-linux-gnuabi64-g++-9
			;;
		windows/386)
			typesuffix=".exe"
			cc=i686-w64-mingw32-gcc-posix
			cxx=i686-w64-mingw32-g++-posix
			# https://docs.microsoft.com/en-us/cpp/porting/modifying-winver-and-win32-winnt?view=vs-2019
			win32_winnt="-D_WIN32_WINNT=0x0A00" # Windows 10
			cgo_cflags="$cgo_cflags $win32_winnt"
			cgo_cxxflags="$win32_winnt"
			;;
		windows/amd64)
			typesuffix=".exe"
			cc=x86_64-w64-mingw32-gcc-posix
			cxx=x86_64-w64-mingw32-g++-posix
			# https://docs.microsoft.com/en-us/cpp/porting/modifying-winver-and-win32-winnt?view=vs-2019
			win32_winnt="-D_WIN32_WINNT=0x0A00" # Windows 10
			cgo_cflags="$cgo_cflags $win32_winnt"
			cgo_cxxflags="$win32_winnt"
			;;
		*)
			echo "$target: Building this os+arch combination is TBD"
			skip=yes
			;;
		esac
	fi
	export \
		CC="$cc" \
		CGO_ENABLED=1 \
		CGO_CFLAGS="$cgo_cflags" \
		CGO_LDFLAGS="$cgo_ldflags" \
		CGO_CXXFLAGS="$cgo_cxxflags" \
		CXX="$cxx" \
		GO111MODULE=on \
		GOARCH="$goarch" \
		GOARM="$goarm" \
		GOOS="$goos" \
		GOPROXY=direct \
		GOSUMDB=off

}

run() {
	check_golang_version
	parse_args "$@"
	if test -z "$action" ; then
		help
		exit 1
	fi
	if test "(" "$action" == build -o "$action" == install ")" -a -z "${targets[*]}" ; then
		help
		exit 1
	fi

	set_go_flags default
	case "$action" in
	build) ## Build apps
		: # handled below
		;;
	coverage) ## Calculate coverage
		c=.testCoverage.txt
		go list ./... | grep -v '/gateway' | xargs go test -covermode=count -coverprofile="$c" && \
			go tool cover -func="$c" && \
			go tool cover -html="$c" -o .testCoverage.html
		return $?
		;;
	gqlgen) ## Run gqlgen
		pushd ./gateway/graphql/ 1>/dev/null || return 1
		go run github.com/99designs/gqlgen --config gqlgen.yml
		code="$?"
		popd 1>/dev/null || return 1
		return "$code"
		;;
	install) ## Build apps (in $GOPATH/bin)
		: # handled below
		;;
	integrationtest) ## Run integration tests (godog)
		go test -v ./integration/... -godog.format=pretty
		return "$?"
		;;
	spec_feature_test) ## Run qa integration tests (godog)
		local repo="${specsrepo:?}"
		if test -z "$repo"; then
			echo "specsrepo not specified"
			exit 1
		fi
		local features
		features="${repo}/qa-scenarios"
		echo "features = $features"
		go test -v ./integration/... --features="$features" -godog.format=pretty
		return "$?"
		;;
	mocks) ## Generate mocks
		go generate ./...
		return "$?"
		;;
	test) ## Run go test
		go test ./...
		return "$?"
		;;
	race) ## Run go test -race
		go test -race ./...
		return "$?"
		;;
	retest) ## Run go test (and force re-test)
		go test -count=1 ./...
		return "$?"
		;;
	buflint) ## Run
		buf lint
		return "$?"
		;;
	misspell) ## Run misspell
		# Since misspell does not support exluding, we need to specify the
		# files we want and those we don't
		find . -name vendor -prune -o "(" -type f -name '*.go' -o -name '*.proto' ")" -print0 | xargs -0 misspell -j 0 -error
		return "$?"
		;;
	semgrep) ## Run semgrep
		semgrep -f "p/dgryski.semgrep-go"
		return "$?"
		;;
	*)
		echo "Invalid action: $action"
		return 1
	esac

	failed=0
	for target in "${targets[@]}" ; do
		set_go_flags "$target"
		if test "$skip" == yes ; then
			continue
		fi
		can_build "$target" "$cc" "$cxx" || continue

		log="/tmp/go.log"
		echo "$target: go mod download ... "
		go mod download 1>"$log" 2>&1
		code="$?"
		if test "$code" = 0 ; then
			echo "$target: go mod download OK"
		else
			echo "$target: go mod download failed ($code)"
			failed=$((failed+1))
			echo
			echo "=== BEGIN logs ==="
			cat "$log"
			echo "=== END logs ==="
			rm "$log"
			continue
		fi

		for app in "${apps[@]}" ; do
			if test "$app" = "v2" ; then
				app=vega
			fi

			echo "Building $app"
			case "$action" in
			build)
				o="cmd/$app/$app$osarchsuffix$dbgsuffix$suffix$typesuffix"
				log="$o.log"
				msgprefix="$target: go $action $o ..."
				echo "$msgprefix"
				rm -f "$o" "$log"
				go build -v -gcflags "$gcflags" -o "$o" "./cmd/$app" 1>"$log" 2>&1
				code="$?"
				;;
			install)
				o="not/applicable"
				log="$app$osarchsuffix$dbgsuffix$suffix$typesuffix.log"
				msgprefix="$target: go $action $app ..."
				echo "$msgprefix"
				rm -f "$log"
				go install -v -gcflags "$gcflags" "./cmd/$app" 1>"$log" 2>&1
				code="$?"
				;;
			esac
			if test "$code" = 0 ; then
				echo "$msgprefix OK"
				rm "$log"
			else
				echo "$msgprefix failed ($code)"
				failed=$((failed+1))
				echo
				echo "=== BEGIN logs for $msgprefix ==="
				cat "$log"
				echo "=== END logs for $msgprefix ==="
			fi
		done
	done
	if test "$failed" -gt 0 ; then
		echo "Build failed for $failed apps."
	fi
	return "$failed"
}

# # #

if echo "$0" | grep -q '/build.sh$' ; then
	# being run as a script
	run "$@"
	exit "$?"
fi
