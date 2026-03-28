import matplotlib.pyplot as plt
import numpy as np
import sys

#нам нужно C, M, E, kappa
def read_file(file_names:list):
    T, C, afm, M, E, kappa, af_kappa = [[] for _ in range(len(file_names))], [[] for _ in range(len(file_names))], [[] for _ in range(len(file_names))], [[] for _ in range(len(file_names))], [[] for _ in range(len(file_names))], [[] for _ in range(len(file_names))], [[] for _ in range(len(file_names))]
    for i in range(len(file_names)):
        with open(file_names[i], "r") as f:
            for line in f:
                point = line.rstrip().split(";")
                T[i].append(float(point[0]))
                C[i].append(float(point[4]))
                afm[i].append(float(point[3]))
                M[i].append(float(point[2]))
                E[i].append(float(point[1]))
                kappa[i].append(float(point[5]))
                af_kappa[i].append(float(point[6]))
    return T, C, M, E, kappa, af_kappa, afm

def plot(points:list,temp:list, name:str):
    x_points = np.array(temp[0])

    for i in range(len(points)):
        y_points = np.array(points[i])
        print(y_points)
        plt.plot(x_points, y_points, label=name+str(i), linewidth=1)
    plt.title(name)
    plt.xlabel("T")
    plt.ylabel(name)
    plt.legend()
    plt.savefig(name + '.png', dpi=300, bbox_inches='tight')
    plt.show()

def main():
    try:
        file_names = sys.argv[1:]
        if len(file_names) == 0:
            raise IndexError
    except IndexError:
        print("Указано недостаточное кол-во файлов, укажите файл для ввода данных")
        return
    T, C, M, E, kappa, af_kappa, afm = read_file(file_names)
    plot(C, T, "C")
    plot(M, T, "M")
    plot(afm, T, "Afm")
    plot(E, T, "E")
    plot(kappa, T, "kappa")
    plot(af_kappa, T, "af_kappa")


main()