import time

def timer(func):
    def wrap(*args, **kwargs):
        start = time.perf_counter()
        func()
        end = time.perf_counter()
        print(end-start)
    return wrap
