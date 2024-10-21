

def iter_example():

    for i in range(10):
        if i == 5:
            continue
        yield i

def test_foo():
    for i in iter_example():
        print(i)