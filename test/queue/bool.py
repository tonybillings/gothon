from queue import Queue


_node_: int = 0
_node_count_: int = 0

_sync_main_: callable = lambda n=_node_count_: ()

_booleans_: Queue[bool] = Queue(10)
booleans = []


def add_boolean_until_full():
    current_boolean = False
    while not _booleans_.full():
        ok = _booleans_.put(current_boolean)
        if ok:
            current_boolean = not current_boolean


def remove_boolean_until_empty():
    while not _booleans_.empty():
        num, ok = _booleans_.get()
        if ok:
            booleans.append(num)


if __name__ == '__main__':
    add_boolean_until_full()
    _sync_main_(1)
    _sync_main_()

    if _node_ == 0:
        print(f'size before empty: {_booleans_.qsize()}')
        remove_boolean_until_empty()
        print(f'size after empty: {_booleans_.qsize()}')
        booleans.sort()
        for boolean in booleans:
            print(f'{boolean} ')
