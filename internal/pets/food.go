package pets

type Food int

func (f Food) String() string {

	if f <= 0 {
		return `Starving`
	}

	if f == 1 {
		return `Hungry`
	}

	if f == 2 {
		return `Satisfied`
	}

	return `Full`
}

func (f *Food) Add() {

	*f += 1

	if *f < 0 {
		*f = 0
	}
	if *f > 3 {
		*f = 3
	}
}

func (f *Food) Remove() {

	*f -= 1

	if *f < 0 {
		*f = 0
	}
	if *f > 3 {
		*f = 3
	}
}
