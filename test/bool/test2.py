
class State:
    _b_: bool = False


s = State()


def get_val():
    return False


if __name__ == '__main__':
    decoyStr = 'blah'
    decoyBool = True
    decoyInt = 2
    decoyFloat = 2.2
    decoyStr2: str = 'blah'
    decoyBool2: bool = True
    decoyInt2: int = 2
    decoyFloat2: float = 2.2

    s._b_ = True
    x = False
    if s._b_:
        x = s._b_
        s._b_ = get_val()

    print(x)
