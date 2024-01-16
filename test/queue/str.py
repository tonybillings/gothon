from queue import Queue


_node_: int = 0
_node_count_: int = 0

_sync_main_: callable = lambda n=_node_count_: ()

_strings_: Queue[str] = Queue(10)
strings = []


def add_string_until_full():
    current_string = ''
    while not _strings_.full():
        ok = _strings_.put(current_string)
        if ok:
            current_string += 'A'


def remove_string_until_empty():
    while not _strings_.empty():
        num, ok = _strings_.get()
        if ok:
            strings.append(num)


if __name__ == '__main__':
    add_string_until_full()
    _sync_main_(1)
    _sync_main_()

    if _node_ == 0:
        print(f'size before empty: {_strings_.qsize()}')
        remove_string_until_empty()
        print(f'size after empty: {_strings_.qsize()}')
        strings.sort()
        for string in strings:
            print(f'{string} ')

