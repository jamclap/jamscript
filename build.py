import os
import os.path as p
import subprocess as sp
import time as t


target_dir = "target"
target_file = p.join(target_dir, "jamscript")


def build():
    start = t.perf_counter()
    mkdir(target_dir)
    run(["go", "build", "-o", target_file, "main.go"])
    end = t.perf_counter()
    print(f"Build time: {end - start:.2f} s")


def main():
    build()
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
