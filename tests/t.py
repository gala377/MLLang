import os
import sys
import subprocess
import glob
import tempfile

from pathlib import Path

"""
Build fnk binary somewhere
Parse test files and write parsed test file to the temporary file
Run the binary on the file and compare output
"""


TEST_FILES_EXTENSION = "fnk"
FNK_BUILD_PACKAGE = "cmd/funk"
FNK_BINARY_NAME = "funk"

tests_dir = Path(__file__).parent
root_dir = tests_dir.parent

def run():
    test_files = get_tests(tests_dir)
    print_tests(tests_dir, test_files)
    with tempfile.TemporaryDirectory() as build_dir:
        build_dir_path = Path(build_dir)
        build_fnk(root_dir, build_dir_path)
        fnk_binary = build_dir_path / FNK_BINARY_NAME
        if not (fnk_binary.exists() and fnk_binary.is_file()):
            raise RuntimeError(f"Funk binary {fnk_binary} does not exist")
        run_tests(fnk_binary, test_files)

def get_tests(path: Path) -> list[Path]:
    return list(path.glob(f"**/*.{TEST_FILES_EXTENSION}"))

def print_tests(tests_root: Path, tests: list[Path]):
    print("Found following test files:")
    for test in tests:
        print(f"\t{strip_tests_dir(test)}")

def build_fnk(root: Path, out: Path) -> Path:
    last_dir = os.getcwd()
    os.chdir(root)
    try:
        print(f"Building latest interpreter binary from {root}\n")
        subprocess.run(
            ["go", "build", f"-o={out}", f"./{FNK_BUILD_PACKAGE}"],
            check=True)
        print("Built...")
    finally:
        os.chdir(last_dir)

def run_tests(binary: Path, tests: list[Path]):
    failed = []
    for test in tests:
        print(f"Running test {strip_tests_dir(test)}... ", end='')
        try:
            subprocess.run([binary, test], check=True, capture_output=True)
        except subprocess.CalledProcessError as e:
            failed.append((test, e.stdout, e.stderr))
            print("Failed")
        else:
            print("Passed")
    print_stats(failed, tests)

def print_stats(failed: list[tuple[Path, bytes, bytes]], all: list[Path]):
    print("="*70)
    print("TEST RESULTS")
    print("="*70)
    for (path, stdout, stderr) in failed:
        print(f"FAILED TEST {strip_tests_dir(path)}")
        print(stdout.decode("utf-8"))
        print(stderr.decode("utf-8"))
        print("\n")

    print(f"Passed {len(all)-len(failed)}/{len(all)}.")
    print(f"{len(failed)} tests failed.")

def strip_tests_dir(path) -> str:
    return str(path)[len(str(tests_dir))+1:]

if __name__ == '__main__':
    run()
