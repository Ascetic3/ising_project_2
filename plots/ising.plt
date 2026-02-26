reset

if (!exists("file")) file = "results/results_J_1_h_0.csv"

set datafile separator ";"
set term pngcairo enhanced size 800,600

set xlabel "T"
set ylabel "<E>/N"
set grid
set key top left
set output sprintf("%s_E.png", file)
plot file using 1:2 with lines lw 2 title "U(T)"

set ylabel "C/N"
set output sprintf("%s_C.png", file)
plot file using 1:4 with lines lw 2 title "C(T)"

set ylabel "m"
set output sprintf("%s_M.png", file)
plot file using 1:3 with lines lw 2 title "m(T)"


