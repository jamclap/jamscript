import os
import os.path as p
import subprocess as sp
import sys
import time as t


target_dir = "target"
target_file = p.join(target_dir, "jams")


def build():
    start = t.perf_counter()
    mkdir(target_dir)
    failed = False
    try:
        extra = []
        extra = ["-ldflags=-s -w"]
        run(["go", "build", *extra, "-o", target_file, "main.go"])
    except:
        failed = True
        pass
    end = t.perf_counter()
    print(f"Build time: {end - start:.2f} s")
    return not failed


def main():
    if not build():
        sys.exit(1)
    report()


def mkdir(path: str):
    os.makedirs(path, exist_ok=True)


def report():
    report_size(target_file)


def report_size(path: str):
    size = p.getsize(path)
    print(f"Size of {path}: {size} B or {size / (1 << 20):.2f} MiB")


def run(cmd, **kwargs):
    sp.run(cmd, check=True, **kwargs)


main()
