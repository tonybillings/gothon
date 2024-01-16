from queue import Queue


_node_: int = 0
_node_count_: int = 0

_sync_main_: callable = lambda n=_node_count_: ()

_numbers_: Queue[int] = Queue(10)
numbers = []


def add_number_until_full():
    current_number = 0
    while not _numbers_.full():
        ok = _numbers_.put(current_number)
        if ok:
            current_number += 1


def remove_number_until_empty():
    while not _numbers_.empty():
        num, ok = _numbers_.get()
        if ok:
            numbers.append(num)


if __name__ == '__main__':
    add_number_until_full()
    _sync_main_(1)
    _sync_main_()

    if _node_ == 0:
        print(f'size before empty: {_numbers_.qsize()}')
        remove_number_until_empty()
        print(f'size after empty: {_numbers_.qsize()}')
        numbers.sort()
        for number in numbers:
            print(f'{number} ')
