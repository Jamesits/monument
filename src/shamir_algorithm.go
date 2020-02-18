package main

import (
	"github.com/SSSaaS/sssa-golang"
)

/*
Shamir algo total keys a, required b
People involved total x, required y (assume everyone have 1 key)
death switch have z keys

constraints:
x+z<=a
y+z>=b
x<b
y<b
z<b
a>=b>0
x>=y>0
z>0

So we have: (in Wolfram language)
Reduce[{x + z <= a, x < b, y + z >= b, x < b, z < b, a >= b, x >= y, b > 0, y > 0, z > 0}, {a, b, x, y, z}, Integers]

which resulted in:
a = 2 n_1 + n_2 + 2 n_3 + n_4 + n_5 + 2, b = n_1 + n_2 + n_3 + n_4 + 2, x = n_1 + n_2 + n_3 + 1, y = n_1 + n_2 + 1, z = n_1 + n_3 + n_4 + 1, n_5 element Z, n_5 >=0, n_4 element Z, n_4 >=0, n_3 element Z, n_3 >=0, n_2 element Z, n_2 >=0, n_1 element Z, n_1 >=0

set n_2, n_4 and n_5 = 0, then we have:
a = 2n_1 + 2n_3 + 2
b = n1 + n3 + 2
x = n1 + n3 + 1
y = n1 + 1
z = n1 + n3 + 1
 */

func getPeopleCount(totalPeople int, requiredPeople int) (totalShares int, requiredShares int, deathSwitchShares int) {
	// TODO: sanity checks

	n3 := totalPeople - requiredPeople
	n1 := requiredPeople - 1
	totalShares = 2*n1 + 2*n3 + 2
	requiredShares = n1 + n3 + 2
	deathSwitchShares = n1 + n3 + 1

	return
}

func shamirEncrypt(data string, totalShares int, requiredShares int) []string {
	ret, err := sssa.Create(requiredShares, totalShares, data)
	hardFailIf(err)

	return ret
}

func shamirDecrypt(collectedShares []string) (string, error) {
	ret, err := sssa.Combine(collectedShares)

	return ret, err
}