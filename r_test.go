// Copyright 2014 Volker Dobler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package export

import (
	"flag"
	"fmt"
	"os/exec"
	"testing"
)

var doR = flag.Bool("R", false, "perform plotting with R")

type Diamond struct {
	Carat   float32
	Cut     string
	Color   string
	Clarity Clarity
	Depth   float32
	Table   float32
	Price   float32
	X       float32
	Y       float32
	Z       float32
}

type Clarity uint8

func (c Clarity) String() string {
	return []string{"FL", "IF", "VVS1", "VVS2", "VS1", "VS2", "SI1", "SI2", "I1", "I2"}[int(c)]
}

var diamonds = []Diamond{
	{1.05, "Ideal", "F", Clarity(3), 61.5, 57, 9138, 6.51, 6.57, 4.02},
	{0.41, "Premium", "J", Clarity(4), 62.1, 58, 647, 4.75, 4.79, 2.96},
	{1.01, "Premium", "H", Clarity(6), 61.8, 60, 4943, 6.44, 6.38, 3.96},
	{1.51, "Premium", "H", Clarity(7), 60.4, 59, 7864, 7.3, 7.27, 4.4},
	{0.9, "Very Good", "I", Clarity(5), 63, 60, 3709, 6.15, 6.17, 3.88},
	{0.53, "Ideal", "D", Clarity(6), 61.7, 56, 1631, 5.15, 5.23, 3.2},
	{0.4, "Very Good", "D", Clarity(6), 61.4, 56, 772, 4.72, 4.83, 2.93},
	{0.32, "Ideal", "E", Clarity(5), 61.6, 55, 702, 4.41, 4.42, 2.72},
	{0.52, "Premium", "F", Clarity(5), 58.5, 60, 1651, 5.29, 5.24, 3.08},
	{1.54, "Premium", "E", Clarity(7), 59, 62, 8020, 7.59, 7.56, 4.47},
	{0.33, "Ideal", "E", Clarity(5), 60.5, 55, 738, 4.52, 4.56, 2.74},
	{0.36, "Ideal", "H", Clarity(3), 61.4, 55, 689, 4.57, 4.6, 2.82},
	{0.31, "Good", "E", Clarity(6), 63.4, 57, 544, 4.31, 4.33, 2.74},
	{0.35, "Premium", "G", Clarity(1), 60.6, 60, 1044, 4.54, 4.57, 2.76},
	{0.55, "Ideal", "F", Clarity(4), 61, 55, 1857, 5.29, 5.33, 3.24},
	{0.4, "Ideal", "G", Clarity(6), 61.5, 57, 942, 4.74, 4.76, 2.92},
	{0.35, "Ideal", "I", Clarity(2), 61.8, 55, 627, 4.53, 4.56, 2.81},
	{0.33, "Premium", "I", Clarity(3), 61, 59, 579, 4.41, 4.44, 2.7},
	{0.33, "Ideal", "F", Clarity(3), 62.4, 55, 750, 4.43, 4.45, 2.77},
	{0.33, "Premium", "H", Clarity(4), 59.7, 59, 743, 4.53, 4.52, 2.7},
	{0.48, "Good", "G", Clarity(4), 58.2, 59.2, 1228, 5.09, 5.14, 2.98},
	{0.41, "Very Good", "G", Clarity(2), 60, 59, 1151, 4.83, 4.84, 2.9},
	{1.01, "Good", "E", Clarity(7), 57.5, 63, 4171, 6.56, 6.58, 3.78},
	{0.35, "Very Good", "G", Clarity(3), 62.8, 58, 798, 4.44, 4.48, 2.8},
	{1.05, "Ideal", "J", Clarity(7), 62.3, 54, 4126, 6.52, 6.55, 4.07},
	{0.72, "Premium", "I", Clarity(6), 61.5, 59, 2262, 5.78, 5.73, 3.54},
	{1.01, "Premium", "I", Clarity(6), 58.6, 62, 4072, 6.51, 6.45, 3.8},
	{0.38, "Ideal", "E", Clarity(5), 61.7, 55, 866, 4.66, 4.68, 2.88},
	{0.65, "Very Good", "H", Clarity(2), 61.6, 58, 2318, 5.57, 5.62, 3.42},
	{0.51, "Ideal", "G", Clarity(4), 62.2, 55, 1656, 5.11, 5.14, 3.19},
	{1.56, "Premium", "E", Clarity(7), 62.4, 60, 10090, 7.41, 7.36, 4.61},
	{0.5, "Good", "F", Clarity(5), 63.6, 57, 1436, 5.03, 5.07, 3.21},
	{0.31, "Good", "D", Clarity(6), 63.5, 57, 571, 4.28, 4.32, 2.73},
	{0.82, "Very Good", "I", Clarity(4), 63.6, 58.9, 2583, 5.88, 5.93, 3.76},
	{0.72, "Ideal", "H", Clarity(7), 61.3, 55, 2410, 5.78, 5.81, 3.55},
	{0.4, "Ideal", "G", Clarity(3), 61.9, 55, 931, 4.72, 4.75, 2.93},
	{0.51, "Ideal", "E", Clarity(4), 62.8, 57, 2075, 5.1, 5.07, 3.2},
	{1.03, "Ideal", "D", Clarity(3), 62, 56, 9798, 6.46, 6.55, 4.03},
	{1.1, "Premium", "H", Clarity(6), 63, 58, 4916, 6.6, 6.55, 4.14},
	{0.56, "Very Good", "J", Clarity(5), 60.6, 57, 1232, 5.35, 5.38, 3.25},
	{0.56, "Ideal", "D", Clarity(2), 61.8, 56, 3270, 5.28, 5.31, 3.27},
	{0.32, "Good", "H", Clarity(3), 63.4, 56, 645, 4.34, 4.37, 2.76},
	{0.6, "Ideal", "G", Clarity(4), 61.3, 60, 2091, 5.39, 5.41, 3.31},
	{0.31, "Ideal", "F", Clarity(1), 62.3, 57, 1122, 4.34, 4.3, 2.69},
	{1.08, "Ideal", "E", Clarity(6), 62.6, 57, 5189, 6.56, 6.53, 4.1},
	{0.39, "Very Good", "G", Clarity(7), 63.2, 57, 597, 4.65, 4.57, 2.92},
	{0.52, "Ideal", "D", Clarity(5), 62, 53, 1701, 5.17, 5.21, 3.22},
	{1.09, "Ideal", "G", Clarity(8), 61, 56, 3549, 6.62, 6.66, 4.05},
	{0.34, "Very Good", "F", Clarity(7), 61.7, 61, 447, 4.44, 4.48, 2.75},
	{1.01, "Ideal", "I", Clarity(7), 62.5, 57, 4336, 6.38, 6.45, 4.01},
	{1, "Good", "D", Clarity(4), 62.7, 62, 7672, 6.21, 6.23, 3.9},
	{0.3, "Ideal", "F", Clarity(2), 61.9, 58, 814, 4.26, 4.3, 2.65},
	{0.79, "Very Good", "F", Clarity(6), 58.5, 61, 3073, 6.03, 6.08, 3.54},
	{0.71, "Premium", "H", Clarity(1), 60.2, 61, 3384, 5.8, 5.76, 3.48},
	{0.26, "Very Good", "E", Clarity(6), 62, 54, 384, 4.08, 4.11, 2.54},
	{1.13, "Very Good", "F", Clarity(6), 61.2, 59, 5017, 6.72, 6.75, 4.12},
	{0.44, "Ideal", "H", Clarity(6), 61.9, 57, 733, 4.9, 4.92, 3.04},
	{0.78, "Ideal", "H", Clarity(6), 61.1, 57, 2616, 5.9, 5.94, 3.62},
	{0.41, "Ideal", "F", Clarity(3), 61.7, 57, 1115, 4.73, 4.8, 2.94},
	{0.51, "Ideal", "G", Clarity(3), 62, 57, 1750, 5.1, 5.13, 3.17},
	{0.31, "Premium", "H", Clarity(4), 62.4, 58, 544, 4.32, 4.36, 2.71},
	{0.41, "Premium", "D", Clarity(3), 62.2, 58, 1181, 4.8, 4.78, 2.98},
	{0.42, "Ideal", "D", Clarity(4), 61.5, 54, 1235, 4.84, 4.79, 2.96},
	{0.9, "Good", "D", Clarity(7), 64, 59, 4078, 6.04, 6.09, 3.88},
	{0.72, "Very Good", "G", Clarity(6), 63.6, 56, 2170, 5.69, 5.72, 3.63},
	{0.31, "Ideal", "H", Clarity(2), 62.3, 53, 687, 4.35, 4.38, 2.72},
	{0.34, "Ideal", "H", Clarity(5), 61.2, 55, 517, 4.49, 4.53, 2.76},
	{0.34, "Ideal", "F", Clarity(3), 61.8, 56, 914, 4.49, 4.47, 2.77},
	{1.63, "Premium", "F", Clarity(6), 59.7, 62, 12394, 7.73, 7.65, 4.59},
	{0.55, "Very Good", "D", Clarity(5), 59.8, 57, 1867, 5.31, 5.35, 3.19},
	{0.51, "Good", "D", Clarity(6), 63.4, 59, 1193, 5.07, 5.02, 3.2},
	{1.25, "Premium", "F", Clarity(5), 60.8, 58, 9702, 6.98, 6.93, 4.23},
	{1, "Ideal", "D", Clarity(2), 60.7, 56, 14498, 6.47, 6.54, 3.95},
	{1.07, "Very Good", "F", Clarity(4), 62, 58, 7118, 6.57, 6.59, 4.08},
	{1, "Good", "E", Clarity(6), 63.1, 64, 4435, 6.31, 6.25, 3.96},
	{1.5, "Ideal", "G", Clarity(7), 62, 57, 8580, 7.26, 7.31, 4.52},
	{1.46, "Very Good", "I", Clarity(6), 62.5, 57, 7146, 7.15, 7.19, 4.48},
	{0.3, "Good", "G", Clarity(4), 63.1, 56, 605, 4.24, 4.26, 2.68},
	{1.07, "Ideal", "G", Clarity(5), 61.7, 55, 7577, 6.53, 6.57, 4.04},
	{1.01, "Fair", "G", Clarity(5), 64.9, 56, 4887, 6.27, 6.21, 4.05},
	{0.7, "Very Good", "F", Clarity(6), 61.7, 59, 2313, 5.64, 5.68, 3.49},
	{0.81, "Premium", "F", Clarity(6), 59.4, 60, 2403, 6.11, 6.07, 3.62},
	{1.7, "Good", "F", Clarity(5), 62.2, 56, 17597, 7.54, 7.6, 4.71},
	{1.71, "Premium", "D", Clarity(7), 60.6, 60, 13325, 7.76, 7.69, 4.68},
	{1.2, "Very Good", "I", Clarity(6), 62.6, 59, 5522, 6.75, 6.71, 4.21},
	{0.4, "Ideal", "G", Clarity(6), 60.2, 56, 900, 4.83, 4.77, 2.89},
	{0.3, "Very Good", "H", Clarity(6), 62.9, 57, 432, 4.22, 4.27, 2.67},
	{0.3, "Premium", "G", Clarity(5), 59.3, 59, 630, 4.42, 4.38, 2.61},
	{1.01, "Very Good", "I", Clarity(6), 63.6, 58, 4189, 6.31, 6.36, 4.03},
	{1.21, "Ideal", "H", Clarity(7), 62.1, 55, 5692, 6.87, 6.82, 4.25},
	{1.02, "Premium", "F", Clarity(7), 61, 58, 4043, 6.49, 6.52, 3.97},
	{0.8, "Very Good", "G", Clarity(4), 63.5, 57, 3381, 5.87, 5.91, 3.74},
	{0.91, "Very Good", "F", Clarity(7), 63, 59, 3405, 6.12, 6.17, 3.87},
	{0.9, "Very Good", "E", Clarity(6), 63.1, 58, 4151, 6.12, 6.02, 3.83},
	{0.32, "Ideal", "G", Clarity(4), 61.9, 55, 645, 4.39, 4.43, 2.73},
	{1.01, "Good", "E", Clarity(5), 56.7, 61, 6606, 6.71, 6.66, 3.79},
	{0.5, "Good", "F", Clarity(6), 64.3, 57, 975, 5.03, 4.94, 3.21},
	{1.52, "Premium", "D", Clarity(7), 61.5, 60, 9789, 7.44, 7.39, 4.56},
	{0.36, "Very Good", "E", Clarity(5), 59.9, 55, 631, 4.66, 4.69, 2.8},
	{0.33, "Ideal", "G", Clarity(1), 61.6, 56, 984, 4.45, 4.48, 2.75},
	{1.21, "Good", "E", Clarity(8), 63.3, 63, 3726, 6.67, 6.72, 4.24},
	{0.25, "Very Good", "E", Clarity(2), 60, 56, 575, 4.1, 4.14, 2.47},
	{0.33, "Ideal", "H", Clarity(1), 61.4, 54, 838, 4.47, 4.49, 2.75},
	{0.43, "Ideal", "G", Clarity(5), 62.4, 56, 1093, 4.87, 4.84, 3.02},
	{0.3, "Ideal", "F", Clarity(1), 62, 56, 886, 4.31, 4.33, 2.68},
	{0.57, "Premium", "G", Clarity(5), 59.7, 59, 1608, 5.37, 5.41, 3.22},
	{0.48, "Premium", "F", Clarity(6), 61.1, 61, 1170, 5.05, 5, 3.07},
	{0.35, "Very Good", "G", Clarity(4), 62.5, 55, 706, 4.51, 4.55, 2.83},
	{1, "Good", "G", Clarity(7), 57.4, 60, 3941, 6.63, 6.53, 3.78},
	{1.6, "Premium", "H", Clarity(5), 62.1, 60, 11796, 7.51, 7.44, 4.64},
	{0.31, "Premium", "G", Clarity(3), 60.5, 58, 707, 4.39, 4.43, 2.67},
	{1.01, "Good", "H", Clarity(6), 63.4, 58, 4116, 6.37, 6.41, 4.05},
	{0.43, "Very Good", "F", Clarity(6), 62.9, 59, 792, 4.79, 4.84, 3.03},
	{0.3, "Premium", "J", Clarity(4), 62.6, 60, 394, 4.22, 4.28, 2.66},
	{1.01, "Ideal", "D", Clarity(6), 62, 55, 5702, 6.39, 6.45, 3.98},
	{0.71, "Very Good", "H", Clarity(4), 63.9, 60, 2562, 5.54, 5.6, 3.56},
	{1.7, "Very Good", "H", Clarity(7), 63.8, 55, 9745, 7.47, 7.55, 4.79},
	{0.4, "Very Good", "E", Clarity(5), 62, 60, 879, 4.67, 4.69, 2.9},
	{1.01, "Premium", "F", Clarity(7), 62.7, 59, 4497, 6.4, 6.35, 4},
	{1.03, "Premium", "F", Clarity(6), 61.5, 59, 5087, 6.49, 6.42, 3.97},
	{0.71, "Ideal", "E", Clarity(6), 61.5, 57, 2308, 5.74, 5.78, 3.54},
	{1.19, "Premium", "G", Clarity(5), 61.9, 58, 7389, 6.86, 6.74, 4.21},
	{0.51, "Very Good", "G", Clarity(5), 62.4, 62, 1438, 5.07, 5.09, 3.17},
	{1.08, "Ideal", "H", Clarity(6), 60, 58, 5867, 6.69, 6.66, 4.01},
	{0.35, "Premium", "D", Clarity(7), 62.3, 58, 669, 4.54, 4.48, 2.81},
	{1.5, "Good", "H", Clarity(5), 63.9, 60, 10692, 7.17, 7.22, 4.6},
	{0.41, "Ideal", "D", Clarity(3), 62.3, 57, 1356, 4.76, 4.74, 2.96},
	{0.7, "Good", "F", Clarity(6), 57.9, 63, 2190, 5.8, 5.84, 3.37},
	{0.33, "Ideal", "G", Clarity(5), 62.2, 56, 579, 4.44, 4.47, 2.77},
	{0.42, "Ideal", "D", Clarity(3), 61.7, 55, 1185, 4.79, 4.83, 2.97},
	{0.4, "Ideal", "E", Clarity(6), 63, 57, 882, 4.68, 4.65, 2.94},
	{0.33, "Ideal", "E", Clarity(4), 62.7, 57, 781, 4.38, 4.45, 2.77},
	{0.4, "Ideal", "E", Clarity(2), 62.6, 56, 1333, 4.7, 4.73, 2.95},
	{0.37, "Premium", "F", Clarity(5), 61.2, 58, 746, 4.63, 4.69, 2.85},
	{1.01, "Premium", "G", Clarity(7), 59.1, 59, 4242, 6.55, 6.51, 3.86},
	{0.73, "Ideal", "J", Clarity(4), 62, 53, 2121, 5.78, 5.82, 3.6},
	{0.31, "Ideal", "D", Clarity(5), 61.9, 56, 942, 4.38, 4.34, 2.7},
	{0.43, "Ideal", "E", Clarity(5), 62.2, 57, 981, 4.8, 4.84, 3},
	{0.9, "Good", "I", Clarity(2), 63.6, 58, 4187, 6.1, 6.14, 3.89},
	{0.41, "Ideal", "F", Clarity(3), 60.6, 57, 1192, 4.86, 4.81, 2.93},
	{0.91, "Premium", "E", Clarity(6), 61.9, 61, 3961, 6.14, 6.11, 3.79},
	{1.15, "Ideal", "F", Clarity(3), 62.1, 55, 8743, 6.69, 6.74, 4.17},
	{0.34, "Premium", "J", Clarity(4), 61.7, 58, 574, 4.47, 4.41, 2.74},
	{1.51, "Very Good", "G", Clarity(5), 63.5, 57, 12872, 7.23, 7.19, 4.58},
	{0.38, "Ideal", "I", Clarity(4), 61.5, 53.9, 703, 4.66, 4.7, 2.89},
	{1.01, "Premium", "G", Clarity(3), 60.5, 60, 7665, 6.58, 6.51, 3.96},
	{1.01, "Premium", "I", Clarity(6), 60.8, 61, 4525, 6.47, 6.43, 3.92},
	{0.72, "Premium", "E", Clarity(5), 61.1, 59, 2954, 5.75, 5.8, 3.53},
	{0.41, "Premium", "E", Clarity(6), 61, 61, 930, 4.77, 4.74, 2.9},
	{0.36, "Premium", "E", Clarity(5), 60.3, 62, 1013, 4.59, 4.56, 2.76},
	{1.07, "Ideal", "G", Clarity(5), 62.2, 57, 6040, 6.55, 6.5, 4.06},
	{1.29, "Ideal", "G", Clarity(7), 61.5, 55, 6321, 7.03, 6.99, 4.31},
	{0.3, "Ideal", "D", Clarity(6), 62.4, 54, 508, 4.32, 4.34, 2.7},
	{0.7, "Very Good", "G", Clarity(5), 61.8, 55, 3026, 5.69, 5.74, 3.53},
	{0.7, "Premium", "E", Clarity(5), 59.3, 60, 2818, 5.78, 5.73, 3.41},
	{0.54, "Ideal", "G", Clarity(6), 61.9, 55, 1736, 5.24, 5.26, 3.25},
	{1.51, "Very Good", "G", Clarity(7), 63.2, 55, 8574, 7.32, 7.27, 4.61},
	{1.08, "Premium", "E", Clarity(7), 60.8, 59, 4656, 6.57, 6.62, 4.01},
	{0.53, "Premium", "H", Clarity(5), 59.4, 59, 1428, 5.31, 5.27, 3.14},
	{0.41, "Ideal", "D", Clarity(5), 62.2, 57, 1007, 4.72, 4.76, 2.95},
	{0.71, "Ideal", "D", Clarity(6), 62.4, 57, 2812, 5.69, 5.72, 3.56},
	{0.31, "Premium", "G", Clarity(3), 62.2, 59, 907, 4.36, 4.32, 2.7},
	{1.2, "Premium", "I", Clarity(6), 61.7, 56, 4872, 6.82, 6.76, 4.19},
	{1.51, "Good", "D", Clarity(7), 61.5, 61, 8742, 7.37, 7.42, 4.55},
	{0.34, "Ideal", "G", Clarity(5), 62.7, 55, 596, 4.48, 4.49, 2.81},
	{0.28, "Very Good", "F", Clarity(3), 61.1, 57, 707, 4.24, 4.27, 2.6},
	{1.33, "Premium", "G", Clarity(7), 60.3, 58, 6565, 7.16, 7.19, 4.33},
	{0.7, "Good", "F", Clarity(4), 59.1, 62, 2877, 5.82, 5.86, 3.44},
	{1.25, "Premium", "I", Clarity(5), 61, 59, 6084, 6.98, 6.89, 4.23},
	{0.5, "Premium", "F", Clarity(5), 60.6, 61, 1433, 5.13, 5.1, 3.1},
	{0.56, "Ideal", "H", Clarity(2), 62.1, 54, 1819, 5.28, 5.31, 3.29},
	{1.22, "Ideal", "G", Clarity(3), 62.3, 56, 10221, 6.84, 6.81, 4.25},
	{0.57, "Premium", "G", Clarity(5), 61.1, 59, 1571, 5.39, 5.32, 3.27},
	{0.5, "Premium", "E", Clarity(5), 58.6, 60, 1654, 5.21, 5.16, 3.04},
	{0.77, "Premium", "F", Clarity(6), 60.8, 59, 2856, 5.92, 5.86, 3.58},
	{0.4, "Ideal", "E", Clarity(3), 62.2, 56, 1158, 4.71, 4.78, 2.95},
	{0.55, "Ideal", "I", Clarity(5), 58.5, 62, 1327, 5.4, 5.39, 3.16},
	{0.54, "Good", "F", Clarity(4), 63.7, 55, 2090, 5.16, 5.11, 3.27},
	{0.33, "Ideal", "F", Clarity(2), 61.8, 57, 955, 4.43, 4.47, 2.75},
	{0.32, "Ideal", "F", Clarity(5), 61.4, 55, 781, 4.43, 4.46, 2.73},
	{0.59, "Very Good", "G", Clarity(4), 57.4, 63, 1915, 5.54, 5.58, 3.19},
	{0.61, "Ideal", "D", Clarity(2), 62.3, 54, 3625, 5.4, 5.45, 3.38},
	{0.71, "Premium", "I", Clarity(2), 63, 58, 2572, 5.68, 5.66, 3.57},
	{1.06, "Premium", "G", Clarity(6), 62.2, 57, 5113, 6.53, 6.47, 4.04},
	{0.56, "Very Good", "E", Clarity(5), 61.1, 57, 1770, 5.33, 5.38, 3.27},
	{0.51, "Ideal", "E", Clarity(2), 62.4, 55, 2317, 5.08, 5.12, 3.18},
	{1.27, "Premium", "J", Clarity(7), 60.4, 58, 4660, 7.04, 7, 4.24},
	{1.01, "Good", "F", Clarity(6), 63.6, 58, 4989, 6.37, 6.31, 4.03},
	{1.02, "Premium", "E", Clarity(7), 58.7, 58, 4990, 6.68, 6.61, 3.9},
	{0.53, "Ideal", "F", Clarity(5), 61.7, 57, 1832, 5.23, 5.21, 3.22},
	{0.41, "Ideal", "E", Clarity(6), 62.7, 57, 969, 4.77, 4.74, 2.98},
	{0.9, "Very Good", "G", Clarity(4), 61.3, 61, 4515, 6.15, 6.25, 3.8},
	{0.5, "Very Good", "G", Clarity(4), 61.3, 58, 1592, 5.03, 5.09, 3.1},
	{1.48, "Premium", "H", Clarity(7), 59.7, 59, 6216, 7.46, 7.42, 4.44},
	{0.72, "Ideal", "E", Clarity(6), 60.3, 57, 2847, 5.83, 5.85, 3.52},
	{0.7, "Ideal", "H", Clarity(5), 61.9, 56, 3656, 5.68, 5.73, 3.53},
	{0.72, "Ideal", "F", Clarity(2), 62, 56, 4252, 5.73, 5.75, 3.56},
	{1.24, "Very Good", "H", Clarity(7), 61, 59, 5231, 6.93, 7, 4.25},
	{1.51, "Ideal", "G", Clarity(7), 61.2, 56.1, 9104, 7.39, 7.41, 4.53},
	{0.79, "Very Good", "F", Clarity(6), 62.9, 57, 3519, 5.86, 5.94, 3.71},
	{0.71, "Premium", "J", Clarity(5), 62.8, 61, 1917, 5.71, 5.63, 3.56},
	{0.51, "Ideal", "D", Clarity(5), 61, 57, 1740, 5.15, 5.18, 3.15},
	{0.33, "Premium", "E", Clarity(6), 61.1, 58, 743, 4.47, 4.43, 2.72},
	{0.4, "Good", "I", Clarity(4), 63.5, 55, 687, 4.67, 4.71, 2.98},
	{0.9, "Good", "E", Clarity(6), 63.8, 61, 3332, 6.08, 6.05, 3.87},
	{1, "Premium", "F", Clarity(5), 62.9, 59, 6296, 6.47, 6.4, 4.02},
	{0.32, "Ideal", "D", Clarity(6), 61.5, 56, 589, 4.39, 4.42, 2.71},
	{0.31, "Ideal", "E", Clarity(2), 62.4, 56, 1028, 4.37, 4.35, 2.72},
	{0.65, "Premium", "F", Clarity(5), 59.6, 58, 1970, 5.65, 5.62, 3.36},
	{0.52, "Ideal", "I", Clarity(4), 62.7, 57, 1720, 5.17, 5.14, 3.23},
	{.28, "Premium", "J", Clarity(7), 62.3, 58, 4636, 6.94, 6.89, 4.31},
	{1.33, "Very Good", "J", Clarity(5), 63.9, 57, 5913, 6.91, 6.96, 4.43},
	{1.2, "Very Good", "H", Clarity(6), 62.4, 60, 6199, 6.75, 6.71, 4.2},
	{0.36, "Very Good", "E", Clarity(5), 62.3, 58, 789, 4.51, 4.55, 2.82},
	{1.28, "Premium", "G", Clarity(5), 61.6, 58, 8874, 7.02, 6.98, 4.31},
	{0.81, "Ideal", "F", Clarity(6), 62.3, 55, 3481, 5.96, 6, 3.72},
	{0.63, "Premium", "E", Clarity(6), 58.8, 59, 1846, 5.68, 5.62, 3.32},
	{1.28, "Premium", "I", Clarity(5), 61.7, 60, 6762, 7.05, 6.95, 4.32},
	{1.22, "Premium", "E", Clarity(5), 62.2, 55, 9292, 6.91, 6.85, 4.28},
	{0.32, "Ideal", "F", Clarity(5), 61.1, 57, 828, 4.44, 4.4, 2.7},
	{0.83, "Ideal", "G", Clarity(6), 61.8, 55, 3171, 6.05, 6.09, 3.75},
	{0.34, "Very Good", "H", Clarity(5), 63.3, 56, 689, 4.45, 4.39, 2.8},
	{0.5, "Ideal", "G", Clarity(2), 61.9, 58, 1883, 5.06, 5.09, 3.14},
	{0.9, "Good", "I", Clarity(6), 63.8, 57, 2823, 6.06, 6.13, 3.89},
	{0.74, "Good", "E", Clarity(5), 63.5, 56, 3163, 5.75, 5.8, 3.67},
	{0.4, "Very Good", "F", Clarity(4), 62, 56, 951, 4.71, 4.73, 2.92},
	{0.83, "Very Good", "D", Clarity(1), 59.7, 53, 7889, 6.14, 6.23, 3.69},
	{0.77, "Fair", "I", Clarity(4), 65.1, 57, 2184, 5.65, 5.77, 3.72},
	{0.43, "Ideal", "G", Clarity(4), 61.9, 55, 1008, 4.86, 4.84, 3},
	{1.08, "Premium", "G", Clarity(7), 62.9, 59, 4627, 6.57, 6.53, 4.12},
	{1.17, "Ideal", "F", Clarity(5), 61.8, 55, 7927, 6.74, 6.81, 4.19},
	{1.04, "Ideal", "G", Clarity(5), 62.6, 55, 6290, 6.49, 6.45, 4.05},
	{0.71, "Ideal", "E", Clarity(7), 58.7, 57, 2340, 5.89, 5.86, 3.44},
	{0.4, "Very Good", "E", Clarity(6), 63.2, 58, 882, 4.7, 4.66, 2.96},
	{0.78, "Very Good", "G", Clarity(6), 63, 58, 2721, 5.82, 5.86, 3.68},
	{0.31, "Ideal", "D", Clarity(4), 62, 54, 877, 4.36, 4.35, 2.7},
	{0.71, "Very Good", "J", Clarity(6), 62.8, 57, 1841, 5.64, 5.67, 3.55},
	{1.26, "Premium", "I", Clarity(7), 60.1, 59, 5151, 7, 6.91, 4.18},
	{1.02, "Ideal", "G", Clarity(7), 61, 57, 4890, 6.51, 6.58, 3.99},
	{1.02, "Premium", "G", Clarity(5), 62.9, 58, 5593, 6.41, 6.37, 4.02},
	{0.34, "Ideal", "H", Clarity(5), 62.6, 54, 689, 4.5, 4.45, 2.8},
	{0.41, "Very Good", "J", Clarity(3), 63.1, 60, 754, 4.67, 4.71, 2.96},
	{1.01, "Very Good", "E", Clarity(7), 63.3, 58, 3674, 6.4, 6.31, 4.02},
	{0.92, "Good", "G", Clarity(3), 58.8, 57, 5390, 6.32, 6.37, 3.73},
	{1.26, "Ideal", "F", Clarity(6), 60.6, 56, 6922, 7.05, 6.98, 4.25},
	{0.9, "Ideal", "J", Clarity(4), 62.5, 55, 3175, 6.18, 6.14, 3.85},
	{0.5, "Very Good", "G", Clarity(5), 62.8, 56, 1373, 5.05, 5.08, 3.18},
	{0.3, "Ideal", "I", Clarity(2), 62.3, 55, 709, 4.29, 4.28, 2.67},
	{0.51, "Premium", "D", Clarity(6), 62.2, 58, 1619, 5.13, 5.06, 3.17},
	{0.32, "Ideal", "J", Clarity(1), 62.2, 55, 533, 4.4, 4.44, 2.75},
	{0.34, "Ideal", "G", Clarity(5), 61.3, 55, 765, 4.51, 4.49, 2.76},
	{0.54, "Ideal", "F", Clarity(5), 60.3, 55, 1786, 5.32, 5.26, 3.19},
	{0.57, "Very Good", "I", Clarity(7), 62.1, 57, 1000, 5.29, 5.33, 3.3},
	{0.52, "Ideal", "E", Clarity(5), 61.1, 57, 1689, 5.18, 5.2, 3.17},
	{0.41, "Ideal", "J", Clarity(5), 62.3, 54, 613, 4.77, 4.8, 2.98},
	{0.73, "Premium", "J", Clarity(4), 60.2, 58, 2037, 5.87, 5.82, 3.52},
	{1.28, "Premium", "H", Clarity(6), 59.9, 59, 6580, 7.05, 7.08, 4.23},
	{0.31, "Ideal", "D", Clarity(5), 61.2, 55, 942, 4.39, 4.37, 2.68},
	{0.4, "Very Good", "E", Clarity(5), 59.1, 60, 814, 4.81, 4.84, 2.85},
	{1.06, "Ideal", "H", Clarity(7), 61.7, 60, 4547, 6.53, 6.51, 4.02},
	{0.54, "Ideal", "G", Clarity(7), 62.3, 58, 1133, 5.17, 5.2, 3.23},
	{0.25, "Premium", "E", Clarity(5), 59.7, 61, 525, 4.1, 4.08, 2.44},
	{0.57, "Ideal", "G", Clarity(4), 62.3, 53.7, 1837, 5.32, 5.35, 3.32},
	{0.31, "Ideal", "E", Clarity(2), 61.5, 56, 865, 4.36, 4.39, 2.69},
	{1.51, "Premium", "H", Clarity(4), 62.4, 60, 10939, 7.34, 7.27, 4.56},
	{0.71, "Ideal", "E", Clarity(5), 62, 55, 3421, 5.72, 5.77, 3.56},
	{0.41, "Premium", "G", Clarity(4), 60.5, 58, 899, 4.76, 4.8, 2.89},
	{0.35, "Fair", "F", Clarity(2), 54.6, 59, 1011, 4.85, 4.79, 2.63},
	{0.34, "Premium", "G", Clarity(5), 60.3, 58, 765, 4.53, 4.49, 2.72},
	{0.32, "Ideal", "I", Clarity(1), 60.8, 54, 864, 4.47, 4.44, 2.71},
	{0.8, "Very Good", "D", Clarity(6), 62.9, 60, 3502, 5.89, 5.94, 3.72},
	{0.34, "Ideal", "G", Clarity(5), 62, 55, 765, 4.48, 4.46, 2.77},
	{0.35, "Ideal", "E", Clarity(4), 61.2, 56, 829, 4.53, 4.55, 2.78},
	{1.3, "Ideal", "F", Clarity(3), 62.1, 55, 12629, 6.96, 7.04, 4.35},
	{1.08, "Premium", "D", Clarity(4), 61, 59, 8999, 6.66, 6.61, 4.05},
	{1.1, "Very Good", "G", Clarity(4), 60.3, 62, 6951, 6.72, 6.67, 4.04},
	{0.5, "Very Good", "D", Clarity(7), 58.5, 60, 1074, 5.19, 5.21, 3.04},
	{0.6, "Ideal", "F", Clarity(5), 60.9, 55, 2099, 5.49, 5.51, 3.34},
	{0.32, "Very Good", "F", Clarity(3), 62.3, 57, 778, 4.38, 4.41, 2.73},
	{1.02, "Good", "H", Clarity(7), 63.7, 58, 3884, 6.28, 6.24, 3.99},
	{2, "Good", "J", Clarity(5), 61.4, 63, 13542, 8.01, 8.08, 4.94},
	{0.62, "Ideal", "G", Clarity(4), 61.8, 57, 2141, 5.47, 5.5, 3.39},
	{0.57, "Ideal", "F", Clarity(6), 61.6, 56, 1299, 5.33, 5.36, 3.29},
	{0.41, "Ideal", "F", Clarity(2), 62.3, 57, 1295, 4.73, 4.77, 2.96},
	{0.75, "Premium", "E", Clarity(5), 61.8, 58, 3206, 5.83, 5.85, 3.61},
	{0.53, "Ideal", "F", Clarity(4), 62.4, 57, 1832, 5.18, 5.21, 3.24},
	{0.42, "Good", "F", Clarity(6), 63.1, 56, 722, 4.79, 4.82, 3.03},
	{0.71, "Premium", "D", Clarity(7), 61.7, 59, 2768, 5.71, 5.67, 3.51},
	{1.65, "Premium", "F", Clarity(7), 60.5, 61, 9693, 7.7, 7.65, 4.64},
	{0.42, "Premium", "G", Clarity(4), 60.9, 58, 984, 4.84, 4.81, 2.94},
	{0.71, "Ideal", "F", Clarity(7), 61.6, 56, 2202, 5.68, 5.74, 3.52},
	{0.4, "Very Good", "I", Clarity(2), 60.8, 58, 849, 4.75, 4.83, 2.91},
	{0.7, "Very Good", "E", Clarity(6), 63.1, 58, 2643, 5.61, 5.64, 3.55},
	{0.31, "Ideal", "G", Clarity(2), 62.4, 56, 772, 4.33, 4.36, 2.71},
	{0.36, "Ideal", "F", Clarity(7), 61.2, 56, 450, 4.61, 4.64, 2.83},
	{1.23, "Premium", "G", Clarity(4), 59.6, 60, 8145, 6.94, 6.99, 4.15},
	{0.75, "Premium", "F", Clarity(7), 59.8, 60, 1881, 5.96, 5.92, 3.55},
	{0.3, "Premium", "D", Clarity(5), 60.7, 60, 710, 4.33, 4.37, 2.64},
	{0.4, "Very Good", "J", Clarity(4), 63, 58, 631, 4.66, 4.7, 2.95},
	{0.36, "Ideal", "E", Clarity(5), 61.7, 57, 878, 4.59, 4.55, 2.82},
	{1, "Fair", "D", Clarity(7), 66.5, 59, 3965, 6.24, 6.21, 4.14},
	{1.06, "Good", "F", Clarity(7), 62.2, 61, 4357, 6.47, 6.53, 4.04},
	{0.3, "Ideal", "J", Clarity(2), 60.4, 57, 464, 4.36, 4.38, 2.64},
	{0.34, "Premium", "I", Clarity(2), 62.6, 58, 803, 4.47, 4.45, 2.79},
	{1.21, "Good", "H", Clarity(4), 63.3, 58, 6748, 6.72, 6.77, 4.27},
	{0.51, "Premium", "E", Clarity(7), 62, 57, 1205, 5.13, 5.09, 3.17},
	{0.41, "Ideal", "E", Clarity(4), 61.5, 56, 1076, 4.79, 4.77, 2.94},
	{0.33, "Ideal", "F", Clarity(3), 62.4, 56, 1067, 4.43, 4.41, 2.76},
	{0.71, "Ideal", "G", Clarity(5), 61.9, 57, 2771, 5.73, 5.77, 3.56},
	{0.9, "Very Good", "H", Clarity(7), 59.6, 59, 3250, 6.38, 6.24, 3.76},
	{0.4, "Premium", "E", Clarity(6), 60.8, 62, 882, 4.76, 4.72, 2.88},
	{0.72, "Premium", "D", Clarity(6), 61.4, 59, 2954, 5.71, 5.79, 3.53},
	{0.3, "Premium", "E", Clarity(3), 61.8, 58, 789, 4.26, 4.29, 2.64},
	{1.06, "Very Good", "H", Clarity(5), 62.4, 56, 5480, 6.49, 6.56, 4.07},
	{0.54, "Premium", "E", Clarity(6), 62.2, 58, 1637, 5.26, 5.23, 3.26},
	{1.22, "Premium", "D", Clarity(6), 62.3, 60, 5739, 6.84, 6.8, 4.25},
	{0.4, "Ideal", "G", Clarity(6), 61, 55, 855, 4.76, 4.81, 2.92},
	{1.88, "Very Good", "H", Clarity(5), 63.1, 55, 12339, 7.86, 7.81, 4.94},
	{0.42, "Premium", "H", Clarity(3), 62.4, 57, 984, 4.83, 4.78, 3},
	{1.24, "Ideal", "F", Clarity(4), 62.2, 55, 9333, 6.89, 6.87, 4.28},
	{0.23, "Good", "F", Clarity(5), 63.8, 57, 357, 3.93, 3.84, 2.48},
	{1.5, "Ideal", "I", Clarity(1), 61.3, 56, 12725, 7.34, 7.41, 4.52},
	{0.49, "Premium", "E", Clarity(4), 58.1, 62, 1577, 5.2, 5.13, 3},
	{0.31, "Ideal", "G", Clarity(3), 62.4, 54, 687, 4.37, 4.38, 2.73},
	{0.4, "Very Good", "D", Clarity(6), 63.3, 57, 924, 4.71, 4.68, 2.97},
	{1.1, "Ideal", "G", Clarity(3), 61.7, 56, 9148, 6.67, 6.63, 4.1},
	{0.33, "Ideal", "H", Clarity(2), 61.7, 56, 752, 4.46, 4.48, 2.76},
	{0.32, "Ideal", "E", Clarity(3), 61.6, 57, 842, 4.4, 4.43, 2.72},
	{0.7, "Ideal", "G", Clarity(4), 61.2, 57, 3140, 5.7, 5.74, 3.5},
	{1.24, "Ideal", "I", Clarity(7), 61.9, 57, 5221, 6.87, 6.92, 4.27},
	{1.5, "Ideal", "G", Clarity(5), 62, 55, 13135, 7.33, 7.36, 4.56},
	{0.57, "Ideal", "E", Clarity(6), 61.9, 57, 2257, 5.35, 5.31, 3.3},
	{1.25, "Premium", "H", Clarity(7), 61.2, 62, 5740, 6.95, 6.84, 4.22},
	{0.7, "Very Good", "E", Clarity(3), 61.5, 58, 3429, 5.67, 5.71, 3.5},
	{0.56, "Very Good", "E", Clarity(5), 61.4, 61, 1819, 5.28, 5.31, 3.25},
	{0.3, "Very Good", "G", Clarity(4), 62.9, 60, 605, 4.26, 4.29, 2.69},
	{0.7, "Very Good", "F", Clarity(7), 62.3, 58, 2070, 5.65, 5.71, 3.54},
	{1.51, "Very Good", "J", Clarity(6), 62.3, 59, 7637, 7.22, 7.28, 4.52},
	{1.08, "Premium", "G", Clarity(4), 63, 59, 7523, 6.56, 6.5, 4.11},
	{1.01, "Premium", "F", Clarity(5), 61.5, 58, 6244, 6.41, 6.46, 3.96},
	{1, "Good", "G", Clarity(4), 63.9, 55, 6028, 6.34, 6.28, 4.03},
	{1.01, "Very Good", "F", Clarity(5), 61.1, 56, 5988, 6.49, 6.56, 3.99},
	{0.56, "Very Good", "F", Clarity(4), 61.2, 58, 1929, 5.32, 5.33, 3.26},
	{0.72, "Good", "F", Clarity(6), 63.9, 62, 2188, 5.7, 5.63, 3.62},
	{0.4, "Good", "D", Clarity(6), 63.6, 57, 720, 4.67, 4.7, 2.98},
	{0.4, "Ideal", "E", Clarity(3), 61.6, 56, 1056, 4.78, 4.73, 2.93},
	{0.55, "Ideal", "G", Clarity(1), 61.1, 57, 2383, 5.26, 5.31, 3.23},
	{0.57, "Ideal", "E", Clarity(6), 61.2, 57, 1613, 5.37, 5.32, 3.27},
	{1, "Fair", "G", Clarity(5), 69.8, 54, 4435, 6.03, 5.94, 4.18},
	{0.91, "Good", "E", Clarity(6), 60.2, 60, 4523, 6.17, 6.26, 3.74},
	{2, "Very Good", "E", Clarity(7), 62.9, 56, 16064, 7.94, 8, 5.01},
	{0.38, "Ideal", "E", Clarity(5), 60.1, 56, 866, 4.69, 4.72, 2.83},
	{0.83, "Ideal", "E", Clarity(4), 62.2, 56, 5151, 5.99, 6.03, 3.74},
	{0.43, "Ideal", "G", Clarity(4), 61.1, 56, 1008, 4.9, 4.86, 2.98},
	{0.34, "Premium", "D", Clarity(6), 59.7, 59, 803, 4.57, 4.54, 2.72},
	{1.04, "Ideal", "H", Clarity(6), 62.7, 54, 4515, 6.51, 6.47, 4.07},
	{0.9, "Very Good", "E", Clarity(7), 59.5, 61, 3139, 6.24, 6.3, 3.73},
	{1.01, "Good", "G", Clarity(2), 63.1, 59, 7279, 6.33, 6.42, 4.02},
	{0.71, "Very Good", "D", Clarity(4), 62, 61, 3420, 5.71, 5.74, 3.55},
	{1, "Good", "I", Clarity(6), 61, 60, 4496, 6.36, 6.39, 3.89},
	{1.06, "Ideal", "H", Clarity(6), 62.2, 57, 5143, 6.56, 6.49, 4.06},
	{0.73, "Ideal", "E", Clarity(7), 60.9, 57, 2581, 5.83, 5.8, 3.54},
	{0.25, "Ideal", "E", Clarity(2), 62.3, 53, 783, 4.08, 4.11, 2.55},
	{0.5, "Very Good", "D", Clarity(6), 58.5, 60, 1436, 5.13, 5.19, 3.02},
	{1.11, "Ideal", "H", Clarity(6), 61.6, 56, 5456, 6.69, 6.65, 4.11},
	{0.3, "Ideal", "F", Clarity(4), 61.8, 55, 612, 4.31, 4.33, 2.67},
	{0.71, "Ideal", "E", Clarity(3), 62, 53.9, 4042, 5.73, 5.77, 3.57},
	{1.15, "Premium", "J", Clarity(6), 62.3, 55, 4347, 6.76, 6.66, 4.18},
	{0.32, "Ideal", "G", Clarity(3), 60.6, 57, 730, 4.41, 4.43, 2.68},
	{0.31, "Premium", "H", Clarity(2), 61.8, 58, 707, 4.3, 4.34, 2.67},
	{1.91, "Premium", "I", Clarity(6), 59.6, 60, 13092, 8.02, 7.98, 4.77},
	{1.03, "Very Good", "E", Clarity(7), 62.1, 59, 4668, 6.43, 6.46, 4},
	{0.71, "Good", "D", Clarity(6), 64, 56, 2641, 5.59, 5.63, 3.59},
	{0.81, "Ideal", "E", Clarity(7), 60.2, 57, 2938, 6.1, 6.06, 3.66},
	{0.56, "Ideal", "G", Clarity(6), 61.6, 59, 1755, 5.27, 5.32, 3.26},
	{1.21, "Premium", "G", Clarity(2), 60.1, 61, 9003, 6.92, 6.85, 4.14},
	{1, "Ideal", "E", Clarity(6), 62.3, 55, 5396, 6.41, 6.34, 3.97},
	{1, "Good", "J", Clarity(6), 58.7, 62, 3614, 6.47, 6.51, 3.81},
	{1.75, "Very Good", "G", Clarity(5), 62.8, 59, 16073, 7.53, 7.62, 4.76},
	{0.33, "Premium", "F", Clarity(5), 62.2, 58, 666, 4.43, 4.44, 2.76},
	{0.8, "Ideal", "I", Clarity(4), 62.2, 58, 2906, 5.92, 5.95, 3.69},
	{1.06, "Ideal", "I", Clarity(7), 61.9, 59, 4103, 6.47, 6.52, 4.02},
	{1, "Good", "E", Clarity(5), 64, 58, 5376, 6.24, 6.2, 3.99},
	{1.54, "Ideal", "J", Clarity(4), 62.2, 59, 8848, 7.34, 7.38, 4.58},
	{0.72, "Very Good", "E", Clarity(5), 62.9, 57, 2990, 5.68, 5.73, 3.59},
	{1, "Good", "E", Clarity(5), 60.6, 65, 6445, 6.29, 6.36, 3.83},
	{0.42, "Ideal", "F", Clarity(3), 62.1, 55, 1221, 4.81, 4.79, 2.98},
	{1, "Good", "H", Clarity(5), 63.7, 59, 4861, 6.3, 6.26, 4},
	{0.34, "Premium", "E", Clarity(5), 61.1, 60, 956, 4.48, 4.45, 2.73},
	{1.24, "Ideal", "I", Clarity(6), 62.1, 56, 5797, 6.84, 6.88, 4.26},
	{0.46, "Premium", "I", Clarity(7), 61, 58, 863, 5.03, 4.97, 3.05},
	{0.3, "Ideal", "E", Clarity(5), 62.6, 54, 658, 4.28, 4.31, 2.69},
	{1.31, "Premium", "G", Clarity(3), 59.6, 61, 11255, 7.23, 7.14, 4.28},
	{0.7, "Very Good", "J", Clarity(6), 61.1, 60, 1959, 5.66, 5.7, 3.47},
	{1.07, "Premium", "E", Clarity(7), 60.5, 59, 4362, 6.67, 6.59, 4.01},
	{0.35, "Ideal", "E", Clarity(3), 61.3, 56, 881, 4.54, 4.6, 2.8},
	{0.53, "Ideal", "E", Clarity(6), 61.4, 56, 1564, 5.2, 5.26, 3.21},
	{0.35, "Premium", "E", Clarity(6), 62.7, 57, 788, 4.52, 4.47, 2.82},
	{0.56, "Ideal", "E", Clarity(4), 61.9, 57, 2145, 5.31, 5.29, 3.28},
	{0.31, "Ideal", "D", Clarity(5), 62.7, 57, 734, 4.3, 4.34, 2.71},
	{0.33, "Premium", "G", Clarity(5), 61.4, 60, 579, 4.41, 4.45, 2.72},
	{0.9, "Premium", "I", Clarity(4), 62.2, 58, 3580, 6.18, 6.13, 3.83},
	{0.43, "Ideal", "F", Clarity(5), 61.1, 56, 968, 4.91, 4.88, 2.99},
	{0.63, "Ideal", "E", Clarity(3), 61.6, 57, 2697, 5.52, 5.49, 3.39},
	{0.9, "Premium", "F", Clarity(7), 61.4, 58, 3619, 6.13, 6.19, 3.78},
	{1, "Ideal", "F", Clarity(5), 61.7, 58, 6424, 6.37, 6.43, 3.95},
	{0.92, "Premium", "J", Clarity(5), 60.9, 62, 3091, 6.31, 6.26, 3.83},
	{1.22, "Ideal", "F", Clarity(4), 62.3, 57, 10100, 6.83, 6.79, 4.24},
	{0.31, "Ideal", "D", Clarity(5), 61.5, 56, 734, 4.34, 4.37, 2.68},
	{0.26, "Very Good", "E", Clarity(2), 63.4, 59, 554, 4, 4.04, 2.55},
	{1, "Very Good", "F", Clarity(6), 59.4, 60, 4961, 6.47, 6.52, 3.86},
	{0.74, "Very Good", "G", Clarity(4), 60.9, 59, 3389, 5.9, 5.86, 3.58},
	{1.36, "Ideal", "F", Clarity(3), 61.6, 56, 12494, 7.08, 7.17, 4.39},
	{0.63, "Good", "F", Clarity(4), 61.6, 60.7, 2188, 5.46, 5.52, 3.38},
	{0.58, "Ideal", "H", Clarity(6), 60.2, 57, 1421, 5.47, 5.43, 3.28},
	{0.31, "Good", "G", Clarity(3), 63.8, 56, 907, 4.28, 4.24, 2.72},
	{0.5, "Ideal", "D", Clarity(4), 61.6, 57, 1923, 5.09, 5.13, 3.15},
	{0.42, "Very Good", "G", Clarity(6), 63.2, 57, 945, 4.79, 4.77, 3.02},
	{0.38, "Ideal", "E", Clarity(6), 61.8, 53.8, 693, 4.63, 4.67, 2.88},
	{1.02, "Very Good", "D", Clarity(7), 62.7, 59, 4712, 6.36, 6.44, 4.01},
	{0.32, "Ideal", "F", Clarity(4), 61.3, 55, 725, 4.4, 4.44, 2.71},
	{0.58, "Ideal", "G", Clarity(4), 62.6, 54, 1984, 5.34, 5.3, 3.33},
	{1.01, "Ideal", "E", Clarity(7), 62.2, 54, 4666, 6.39, 6.43, 3.99},
	{0.71, "Ideal", "J", Clarity(7), 61.3, 56, 1901, 5.73, 5.78, 3.53},
	{0.3, "Very Good", "G", Clarity(3), 63.1, 56, 878, 4.23, 4.2, 2.66},
	{0.31, "Very Good", "E", Clarity(6), 62.8, 57, 544, 4.32, 4.34, 2.72},
	{1.5, "Very Good", "H", Clarity(4), 63.4, 59, 10652, 7.13, 7.2, 4.54},
	{0.32, "Ideal", "F", Clarity(5), 61.3, 56, 645, 4.41, 4.43, 2.71},
	{1.5, "Ideal", "E", Clarity(6), 60.4, 59, 12265, 7.38, 7.43, 4.47},
	{0.54, "Ideal", "G", Clarity(6), 62.4, 56, 1350, 5.2, 5.22, 3.25},
	{1.01, "Very Good", "D", Clarity(7), 61.4, 60, 4564, 6.33, 6.38, 3.9},
	{0.51, "Premium", "G", Clarity(3), 61.5, 56, 1974, 5.15, 5.12, 3.16},
	{0.32, "Ideal", "I", Clarity(4), 60.9, 57, 523, 4.43, 4.47, 2.71},
	{0.37, "Premium", "D", Clarity(6), 61.9, 59, 874, 4.65, 4.62, 2.87},
	{0.9, "Very Good", "J", Clarity(4), 62.9, 56, 3197, 6.13, 6.18, 3.87},
	{1.06, "Good", "G", Clarity(5), 63.1, 59, 6212, 6.45, 6.48, 4.08},
	{0.52, "Ideal", "G", Clarity(6), 62.2, 54, 1352, 5.18, 5.2, 3.23},
	{0.23, "Very Good", "D", Clarity(7), 63.5, 57, 449, 3.87, 3.85, 2.45},
	{0.23, "Very Good", "F", Clarity(4), 61.9, 58, 402, 3.9, 4.02, 2.45},
	{0.3, "Premium", "E", Clarity(5), 62.1, 58, 844, 4.31, 4.29, 2.67},
	{1.05, "Ideal", "H", Clarity(6), 62.4, 57, 5128, 6.49, 6.52, 4.06},
	{0.72, "Very Good", "G", Clarity(7), 62.1, 59, 2015, 5.66, 5.73, 3.54},
	{0.3, "Very Good", "I", Clarity(4), 61.8, 63, 442, 4.26, 4.29, 2.64},
	{1.07, "Premium", "G", Clarity(7), 62.2, 58, 3545, 6.56, 6.49, 4.06},
	{0.33, "Ideal", "F", Clarity(5), 61.6, 56, 666, 4.44, 4.46, 2.74},
	{0.35, "Ideal", "F", Clarity(5), 62.6, 56, 906, 4.52, 4.49, 2.82},
	{0.31, "Very Good", "J", Clarity(5), 62.3, 60, 380, 4.29, 4.34, 2.69},
	{0.31, "Ideal", "I", Clarity(6), 60.9, 57, 414, 4.36, 4.41, 2.67},
	{0.84, "Very Good", "D", Clarity(4), 60.6, 57, 4443, 6.06, 6.15, 3.7},
	{1.01, "Premium", "E", Clarity(6), 58.2, 59, 5366, 6.59, 6.54, 3.82},
	{0.31, "Premium", "H", Clarity(7), 59.5, 59, 390, 4.39, 4.45, 2.63},
	{1.01, "Ideal", "E", Clarity(8), 62, 57, 3450, 6.41, 6.37, 3.96},
	{0.35, "Ideal", "D", Clarity(6), 61.1, 56, 780, 4.55, 4.59, 2.79},
	{0.7, "Very Good", "G", Clarity(5), 62.2, 61, 2633, 5.54, 5.59, 3.46},
	{0.33, "Ideal", "E", Clarity(5), 61, 57, 928, 4.46, 4.43, 2.71},
	{1.22, "Good", "H", Clarity(7), 63.5, 56, 6250, 6.84, 6.77, 4.32},
	{0.5, "Premium", "G", Clarity(6), 63, 56, 1138, 5.07, 5.02, 3.18},
	{0.71, "Premium", "D", Clarity(2), 58.8, 58, 3952, 5.89, 5.81, 3.44},
	{2, "Fair", "H", Clarity(7), 55.3, 65, 11202, 8.42, 8.35, 4.64},
	{0.9, "Good", "I", Clarity(4), 62.8, 57, 3398, 6.07, 6.16, 3.84},
	{0.34, "Premium", "F", Clarity(6), 59.3, 60, 727, 4.58, 4.52, 2.7},
	{0.7, "Very Good", "F", Clarity(3), 60.5, 60, 3205, 5.7, 5.73, 3.46},
	{1.06, "Very Good", "G", Clarity(8), 60.7, 63, 2080, 6.59, 6.56, 3.99},
	{0.8, "Good", "E", Clarity(4), 60.2, 54, 4444, 6.04, 6.01, 3.63},
	{1.51, "Very Good", "J", Clarity(4), 63.7, 55, 7953, 7.16, 7.28, 4.6},
	{1.23, "Very Good", "F", Clarity(4), 60.6, 58, 9346, 6.88, 6.95, 4.19},
	{0.96, "Fair", "H", Clarity(5), 68.8, 56, 3658, 6.11, 5.98, 4.16},
	{1.05, "Very Good", "G", Clarity(7), 62.5, 60, 4241, 6.45, 6.5, 4.05},
	{1.1, "Premium", "H", Clarity(7), 62.6, 57, 4558, 6.6, 6.56, 4.12},
	{1.31, "Premium", "E", Clarity(6), 62.4, 58, 8767, 7.03, 6.95, 4.36},
	{1.01, "Very Good", "I", Clarity(6), 63.5, 59, 4412, 6.36, 6.31, 4.02},
	{0.7, "Ideal", "I", Clarity(4), 61.9, 56, 2616, 5.69, 5.72, 3.53},
	{0.6, "Premium", "H", Clarity(6), 62, 56, 1433, 5.44, 5.4, 3.36},
	{1.23, "Ideal", "G", Clarity(7), 58.8, 60, 6005, 7.01, 7.08, 4.14},
	{0.9, "Very Good", "D", Clarity(7), 60.6, 61, 3534, 6.14, 6.2, 3.74},
	{1.5, "Premium", "H", Clarity(5), 61.8, 59, 11360, 7.3, 7.35, 4.53},
	{0.71, "Premium", "H", Clarity(7), 62.4, 56, 2028, 5.7, 5.64, 3.54},
	{0.53, "Ideal", "F", Clarity(7), 62.4, 54, 1078, 5.19, 5.22, 3.25},
	{0.46, "Very Good", "G", Clarity(4), 62.8, 57, 1049, 4.92, 4.83, 3.06},
	{0.3, "Ideal", "F", Clarity(1), 61.3, 55, 873, 4.32, 4.36, 2.66},
	{0.51, "Good", "E", Clarity(6), 63.9, 56, 1343, 5.07, 5.11, 3.25},
	{0.33, "Premium", "E", Clarity(6), 60.6, 60, 743, 4.44, 4.41, 2.68},
	{0.63, "Good", "J", Clarity(6), 59.9, 64, 1134, 5.63, 5.58, 3.36},
	{0.36, "Very Good", "E", Clarity(6), 62.4, 55, 631, 4.52, 4.55, 2.83},
	{1.51, "Ideal", "I", Clarity(7), 61.2, 58, 7362, 7.33, 7.42, 4.51},
	{2, "Ideal", "H", Clarity(8), 62.5, 57, 7204, 8.05, 7.98, 5.01},
	{1.07, "Premium", "H", Clarity(5), 62.1, 59, 5327, 6.52, 6.56, 4.06},
	{0.55, "Ideal", "I", Clarity(5), 61.8, 55, 1374, 5.28, 5.32, 3.27},
	{0.43, "Premium", "F", Clarity(5), 59.8, 58, 968, 4.94, 4.89, 2.94},
	{0.34, "Premium", "E", Clarity(6), 61.9, 59, 765, 4.48, 4.44, 2.76},
	{1.25, "Very Good", "H", Clarity(7), 61.6, 54, 5637, 6.85, 6.95, 4.25},
	{1.31, "Very Good", "I", Clarity(7), 60.9, 58, 5652, 7.03, 7.07, 4.29},
	{0.58, "Very Good", "J", Clarity(5), 62.8, 57, 1090, 5.28, 5.33, 3.33},
	{1.5, "Good", "G", Clarity(5), 63.3, 62, 11577, 7.08, 7.2, 4.52},
	{1.11, "Ideal", "E", Clarity(6), 61.1, 57, 6800, 6.64, 6.72, 4.08},
	{0.9, "Good", "I", Clarity(7), 64, 55, 2503, 6.11, 6.05, 3.89},
	{1.1, "Ideal", "I", Clarity(4), 61.7, 57, 5544, 6.65, 6.62, 4.09},
	{0.35, "Ideal", "E", Clarity(4), 61.9, 56, 829, 4.51, 4.53, 2.8},
	{0.5, "Ideal", "F", Clarity(3), 61.6, 56, 2160, 5.11, 5.08, 3.14},
	{0.5, "Very Good", "H", Clarity(5), 61.8, 63, 1074, 5.05, 5.08, 3.13},
}

func TestRPlot(t *testing.T) {
	if !*doR {
		t.SkipNow()
	}

	extractor, err := NewExtractor(diamonds, "Carat", "Cut", "Color" /* "Clarity.String", */, "Price")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	args := []string{"--vanilla", "--interactive"}
	cmd := exec.Command("/usr/bin/R", args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	d := RVecDumper{
		Writer: stdin,
		Name:   "my.diamonds",
	}
	d.Dump(extractor, RFormat)

	go func() {
		fmt.Fprintf(stdin, `
library(ggplot2)
p <- ggplot(my.diamonds, aes(Carat, Price))
p + geom_point()
Sys.sleep(2)
`)
		stdin.Close()
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", string(out))
		t.Fatalf("Unexpected error: %s", err)
	}
}