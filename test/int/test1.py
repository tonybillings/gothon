
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

    _i_: int = 10
    _i_ += 5
    _i_ -= 20
    _i_ *= 3
    _i_ /= -2
    x = _i_
    _i_ = get_val()
    _i_ = _i_
    x += _i_

    print(x)
