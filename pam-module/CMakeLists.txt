cmake_minimum_required(VERSION 3.28.0 FATAL_ERROR)

project(ssh-with-passkey LANGUAGES C)

SET(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# c言語のファイルを追加
file(GLOB SRC src/*.c)

add_compile_options("-std=c99" "-Wall" "-Wextra" "-pedantic"
	"-Wstrict-overflow" "-fno-strict-aliasing" "-D _XOPEN_SOURCE=600" "-g"
	"-pedantic" "-pedantic-errors" "-Wformat=2" "-Wnull-dereference"
	"-Winit-self" "-Wmissing-include-dirs" "-Wshift-overflow=2"
	"-Wswitch-default" "-Wswitch-enum"
	"-Wuninitialized" "-Wsuggest-attribute=pure"
	"-Wsuggest-attribute=const" "-Wsuggest-attribute=noreturn"
	"-Wsuggest-attribute=format" "-Wduplicated-cond" "-Wtrampolines"
	"-Wundef" "-Wshadow" "-funsafe-loop-optimizations"
	"-Wunsafe-loop-optimizations" "-Wbad-function-cast" "-Wcast-qual"
	"-Wcast-align" "-Wwrite-strings" "-Wjump-misses-init" "-Wlogical-op"
	"-Waggregate-return" "-Wstrict-prototypes" "-Wold-style-definition"
	"-Wmissing-prototypes" "-Wmissing-declarations" "-Wredundant-decls"
	# 使用しない引数が出るのは当然なのでエラーを出さないようにする
	"-Wno-unused-parameter")

add_library(main SHARED "${SRC}")

target_link_libraries(main pam)

set_target_properties(main PROPERTIES PREFIX "")

set_target_properties(main PROPERTIES SUFFIX ".so")

