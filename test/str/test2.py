
class State:
    _s_: str = 'Hello'


s = State()


def get_val():
    return 'xxx'


if __name__ == '__main__':
    decoyStr = 'blah'
    decoyBool = True
    decoyInt = 2
    decoyFloat = 2.2
    decoyStr2: str = 'blah'
    decoyBool2: bool = True
    decoyInt2: int = 2
    decoyFloat2: float = 2.2

    s._s_ += ' World!!!'
    s._s_ -= '!!'
    x = s._s_
    s._s_ = get_val()

    print(x)
