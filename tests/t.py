import os
import subprocess
import tempfile
import re
import argparse

from dataclasses import dataclass
from pathlib import Path
from typing import Optional

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

def run(args):
    pat = args.filter[0] if args.filter else None
    test_files = get_tests(tests_dir, pattern=pat)
    print_tests(tests_dir, test_files)
    with tempfile.TemporaryDirectory() as build_dir:
        build_dir_path = Path(build_dir)
        build_fnk(root_dir, build_dir_path)
        fnk_binary = build_dir_path / FNK_BINARY_NAME
        if not (fnk_binary.exists() and fnk_binary.is_file()):
            raise RuntimeError(f"Funk binary {fnk_binary} does not exist")
        run_tests(fnk_binary, test_files)

def get_tests(path: Path, pattern: Optional[str]) -> list[Path]:
    files = list(path.glob(f"**/*.{TEST_FILES_EXTENSION}"))
    if pattern is not None:
        return [f for f in files
               if re.match(pattern, f.name)]
    return files

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

@dataclass
class TestFailed(Exception):
    test: Path
    msg: str

def run_tests(binary: Path, tests: list[Path]):
    failed = []
    for test in tests:
        print(f"Running test {strip_tests_dir(test)}... ", end='')
        test_path, expected = parse_test_file(test)
        test_output = expected is not None
        if test_output:
            test_file = test_path
            test_path = test_file.name
        try:
            proc = subprocess.run([binary, test_path], check=True, capture_output=True)
            if test_output:
                output = proc.stdout.splitlines()
                output = list(map(lambda s: s.decode("utf-8"), output))
                fail_msg = (
                    "STDOUT DOES NOT MATCH:\nExpected: ==========\n\n" +
                    "\n".join(expected) +
                    "\nGOT: ===========\n\n" +
                    "\n".join(output))
                if len(expected) != len(output):
                    raise TestFailed(test, fail_msg)
                for e_line, a_line in zip(expected, output):
                    if e_line != a_line:
                        raise TestFailed(test, fail_msg)
        except subprocess.CalledProcessError as e:
            failed.append((
                test,
                e.stdout.decode("utf-8"),
                e.stderr.decode("utf-8")))
            print("Failed")
        except TestFailed as e:
            failed.append((test, e.msg, ""))
            print("Failed")
        else:
            print("Passed")
        finally:
            if test_output:
                test_file.close()
    print_stats(failed, tests)

def parse_test_file(test: Path):
    with test.open() as tf:
        if (l := tf.readline()) != "@EXPECTED\n":
            return test, None
        expected = []
        source = tf.readlines()
        while (line := source.pop(0)) != "@SOURCE\n":
            expected.append(line)
        test_source = tempfile.NamedTemporaryFile()
        test_source.writelines(map(lambda x: bytes(x, 'utf-8'), source))
        test_source.seek(0)
        test_source.flush()
        expected = list(filter(bool, map(str.rstrip, expected)))
        return test_source, expected

def print_stats(failed: list[tuple[Path, bytes, bytes]], all: list[Path]):
    print("="*70)
    print("TEST RESULTS")
    print("="*70)
    for (path, stdout, stderr) in failed:
        print(f"FAILED TEST {strip_tests_dir(path)}")
        print(stdout)
        print(stderr)
        print("\n")

    print(f"Passed {len(all)-len(failed)}/{len(all)}.")
    print(f"{len(failed)} tests failed.")

def strip_tests_dir(path) -> str:
    return str(path)[len(str(tests_dir))+1:]

def command_parser():
    parser = argparse.ArgumentParser(description='tool to run fnk e2e tests')
    parser.add_argument('--filter', type=str, nargs=1,
                        help='only run tests matching the regex')
    return parser

if __name__ == '__main__':
    parser = command_parser()
    args = parser.parse_args()
    run(args)
