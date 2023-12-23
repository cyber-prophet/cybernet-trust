package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
)

type UserOutLink struct {
	UserID  uint64
	OutLink uint64
}

type Pair struct {
	First  uint64
	Second float64
}

func findIntersection(arr1, arr2 []int) []int {
	intersection := make([]int, 0)

	set := make(map[int]bool)

	// Create a set from the first array
	for _, num := range arr1 {
		set[num] = true // setting the initial value to true
	}

	// Check elements in the second array against the set
	for _, num := range arr2 {
		if set[num] {
			intersection = append(intersection, num)
		}
	}

	return intersection
}

func cosineSimilarity(a, b []int) float64 {
	//  var sumAB, sumA2, sumB2 float64
	//  for i := 0; i < len(a); i++ {
	//   sumAB += float64(a[i] * b[i])
	//   sumA2 += float64(a[i] * a[i])
	//   sumB2 += float64(b[i] * b[i])
	//  }
	//  return sumAB / (math.Sqrt(sumA2) * math.Sqrt(sumB2))
	var intersection = findIntersection(a, b)
	// fmt.Printf("Intersection %d and intersection len %d and A %.2f and B %.2f \n", intersection, len(intersection), math.Pow(float64(len(a)),2), math.Pow(float64(len(b)),2))
	return float64(len(intersection)) / math.Sqrt(float64(len(a)*len(b)))
}

func matmulSparse(sparseMatrix [][]Pair, vector []float64, columns uint16) []float64 {
	result := make([]float64, columns)

	for i, sparseRow := range sparseMatrix {
		for _, value := range sparseRow {
			result[value.First] += vector[i] * value.Second
			//fmt.Printf("%d %.2f %.2f \n", value.First, vector[i], value.Second)
		}
	}
	return result
}

func weightedMedianColSparse(stake []float64, score [][]Pair, columns uint16, majority float64) []float64 {
	rows := len(stake)
	zero := float64(0)
	useStake := make([]float64, 0)
	for _, s := range stake {
		if s > zero {
			useStake = append(useStake, s)
		}
	}
	inplaceNormalize(useStake)
	stakeSum := sum(useStake)
	stakeIdx := makeRange(0, len(useStake))
	minority := stakeSum - majority
	useScore := make([][]float64, columns)
	for i := range useScore {
		useScore[i] = make([]float64, len(useStake))
	}
	median := make([]float64, columns)
	k := 0
	for r := 0; r < rows; r++ {
		if stake[r] <= zero {
			continue
		}
		for _, val := range score[r] {
			useScore[val.First][k] = val.Second
		}
		k++
	}
	for c := 0; c < int(columns); c++ {
		median[c] = weightedMedian(useStake, useScore[c], stakeIdx, minority, zero, stakeSum)
	}
	return median
}

func inplaceNormalize(x []float64) {
	xSum := sum(x)
	if xSum == 0 {
		return
	}
	for i := range x {
		x[i] = x[i] / xSum
	}
}

func weightedMedian(stake []float64, score []float64, partitionIdx []int, minority float64, partitionLo float64, partitionHi float64) float64 {
	n := len(partitionIdx)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return score[partitionIdx[0]]
	}
	midIdx := n / 2
	pivot := score[partitionIdx[midIdx]]
	loStake := float64(0)
	hiStake := float64(0)
	lower := make([]int, 0)
	upper := make([]int, 0)
	for _, idx := range partitionIdx {
		if score[idx] == pivot {
			continue
		}
		if score[idx] < pivot {
			loStake += stake[idx]
			lower = append(lower, idx)
		} else {
			hiStake += stake[idx]
			upper = append(upper, idx)
		}
	}
	if partitionLo+loStake <= minority && minority < partitionHi-hiStake {
		return pivot
	} else if minority < partitionLo+loStake && len(lower) > 0 {
		return weightedMedian(stake, score, lower, minority, partitionLo, partitionLo+loStake)
	} else if partitionHi-hiStake <= minority && len(upper) > 0 {
		return weightedMedian(stake, score, upper, minority, partitionHi-hiStake, partitionHi)
	}
	return pivot
}

func sum(a []float64) float64 {
	sum := float64(0)
	for _, v := range a {
		sum += v
	}
	return sum
}

func makeRange(min, max int) []int {
	a := make([]int, max-min)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func colClipSparse(sparseMatrix [][]Pair, colThreshold []float64) [][]Pair {
	result := make([][]Pair, len(sparseMatrix))
	for i, sparseRow := range sparseMatrix {
		for _, value := range sparseRow {
			if colThreshold[value.First] < value.Second {
				if 0 < colThreshold[value.First] {
					result[i] = append(result[i], Pair{value.First, colThreshold[value.First]})
				}
			} else {
				result[i] = append(result[i], value)
			}
		}
	}
	return result
}

func rowSumSparse(sparseMatrix [][]Pair) []float64 {
	rows := len(sparseMatrix)
	result := make([]float64, rows)
	for i, sparseRow := range sparseMatrix {
		for _, value := range sparseRow {
			result[i] += value.Second
		}
	}
	return result
}

func vecdiv(x []float64, y []float64) []float64 {
	if len(x) != len(y) {
		panic("Length of slices x and y must be equal")
	}
	n := len(x)
	result := make([]float64, n)
	for i := 0; i < n; i++ {
		if y[i] != 0 {
			result[i] = x[i] / y[i]
		}
	}
	return result
}

func main() {
	// Assuming you have a slice of UserOutLink
	//userOutLinks := []UserOutLink{
	//	{UserID: 1, OutLink: 1},
	//	{UserID: 1, OutLink: 2},
	//	{UserID: 1, OutLink: 3},
	//	{UserID: 1, OutLink: 4},
	//	{UserID: 2, OutLink: 3},
	//	{UserID: 2, OutLink: 5},
	//	{UserID: 2, OutLink: 6},
	//	{UserID: 3, OutLink: 6},
	//	{UserID: 3, OutLink: 1},
	//	{UserID: 3, OutLink: 5},
	//	{UserID: 3, OutLink: 2},
	//	{UserID: 3, OutLink: 7},
	//	{UserID: 3, OutLink: 8},
	//}
	file, err := os.Open("2_1_neuronid_linkid.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var userOutLinks []UserOutLink

	// Create a user-outlink matrix
	userOutLinkMatrix := make(map[uint64][]int)
	//for _, uol := range userOutLinks {
	//	userOutLinkMatrix[uol.UserID] = append(userOutLinkMatrix[uol.UserID], int(uol.OutLink))
	//}

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		userID, _ := strconv.ParseUint(record[0], 10, 64)
		outLink, _ := strconv.ParseUint(record[1], 10, 64)

		userOutLinks = append(userOutLinks, UserOutLink{UserID: userID, OutLink: outLink})
	}
	//fmt.Printf("userOutLinks %d\n", userOutLinks)
	for _, uol := range userOutLinks {
		userOutLinkMatrix[uol.UserID] = append(userOutLinkMatrix[uol.UserID], int(uol.OutLink))
	}

	sparseWeightsMatrix := make([][]Pair, 1000)

	// Compute cosine similarity between each pair of users
	for userID1, outLinks1 := range userOutLinkMatrix {
		for userID2, outLinks2 := range userOutLinkMatrix {
			if userID1 != userID2 {
				cosim := cosineSimilarity(outLinks1, outLinks2)
				//sparseWeightsMatrix = append(sparseWeightsMatrix, [][]uint64{{uint64(userID1), uint64(userID2), uint64(cosim * 1e6)}})
				sparseWeightsMatrix[userID1] = append(sparseWeightsMatrix[userID1], Pair{userID2, cosim})

				//fmt.Printf("Cosine similarity between user %d and user %d: %.2f\n", userID1, userID2, cosim)
			}
		}
	}

	fmt.Println("Sparse Matrix: ", sparseWeightsMatrix)

	file2, err := os.Open("1_3_neuronid_balance.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file2.Close()
	reader2 := csv.NewReader(file2)

	//stakes := []float64{0, 2000.0, 3000.0, 4000.0}
	stakes := make([]float64, 1000)

	for {
		record, err := reader2.Read()
		if err != nil {
			break
		}

		userID, _ := strconv.ParseUint(record[0], 10, 64)
		stake, _ := strconv.ParseUint(record[1], 10, 64)

		stakes[userID] = float64(stake)
	}

	//fmt.Println("Stakes: ", stakes)
	columns := uint16(len(stakes))

	// Compute preranks: r_j = SUM(i) w_ij * s_i
	preranks := matmulSparse(sparseWeightsMatrix, stakes, columns)
	//fmt.Println("Preranks: ", preranks)

	// Clip weights at majority consensus
	kappa := float64(0.10)
	fmt.Println("Kappa: ", kappa)

	consensus := weightedMedianColSparse(stakes, sparseWeightsMatrix, columns, kappa)
	fmt.Println("Consensus: ", consensus)

	weights := colClipSparse(sparseWeightsMatrix, consensus)

	fmt.Println("Weights: ", weights)

	validator_trust := rowSumSparse(weights)
	fmt.Println("Validator trust: ", validator_trust)

	// Compute ranks: r_j = SUM(i) w_ij * s_i.
	ranks := matmulSparse(weights, stakes, columns)
	fmt.Println("Ranks: ", ranks)

	trust := vecdiv(ranks, preranks)
	fmt.Println("Trust: ", trust)
}
