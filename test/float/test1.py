
def get_val():
    return 1.5


if __name__ == '__main__':
    decoyStr = 'blah'
    decoyBool = True
    decoyInt = 2
    decoyFloat = 2.2
    decoyStr2: str = 'blah'
    decoyBool2: bool = True
    decoyInt2: int = 2
    decoyFloat2: float = 2.2

    _f_: float = 10.5
    _f_ += 5.5
    _f_ -= 20.5
    _f_ *= 3.5
    _f_ /= -2.5
    x = _f_
    _f_ = get_val()
    _f_ = _f_
    x += _f_

    print(x)
