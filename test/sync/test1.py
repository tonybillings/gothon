
_node_: int = 0
_node_count_: int = 0

_lock_file_: callable = lambda: ()
_unlock_file_: callable = lambda: ()
_sync_file_: callable = lambda n=_node_count_: ()


if __name__ == '__main__':
    decoyStr = 'blah'
    decoyBool = True
    decoyInt = 2
    decoyFloat = 2.2
    decoyStr2: str = 'blah'
    decoyBool2: bool = True
    decoyInt2: int = 2
    decoyFloat2: float = 2.2

    _lock_file_()
    with open('/tmp/_gothon_test_counter.dat', mode='a+') as f:
        f.seek(0)
        fileContents = f.read()

    if fileContents == '':
        fileContents = '0'

    counter = int(fileContents)
    counter += 1

    with open('/tmp/_gothon_test_counter.dat', mode='w+') as f:
        f.write(str(counter))
        f.flush()
    _unlock_file_()

    _sync_file_(1)

    if _node_ == 0:
        _sync_file_()
        with open('/tmp/_gothon_test_counter.dat', mode='r') as f:
            fileContents = f.read()
            print(f'total: {fileContents}')
