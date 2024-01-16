

class State:
    _i_: int = 0


s = State()


def get_val():
    return 1


if __name__ == '__main__':
    decoyStr = 'blah'
    decoyBool = True
    decoyInt = 2
    decoyFloat = 2.2
    decoyStr2: str = 'blah'
    decoyBool2: bool = True
    decoyInt2: int = 2
    decoyFloat2: float = 2.2

    s._i_ += 5
    s._i_ -= 20
    s._i_ *= 3
    s._i_ /= -2
    x = s._i_
    s._i_ = get_val()
    s._i_ = s._i_
    x += s._i_

    print(x)
