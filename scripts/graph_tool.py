import matplotlib.pyplot as plt
import numpy as np
import sys


#нам нужно C, M, E, kappa
def read_file(file_name:str):
    T, C, afm, M, E, kappa, af_kappa = [], [], [], [], [], [], []
    with open(file_name, "r") as f:
        for line in f:
            point = line.rstrip().split(";")
            T.append(float(point[0]))
            C.append(float(point[4]))
            afm.append(float(point[3]))
            M.append(float(point[2]))
            E.append(float(point[1]))
            kappa.append(float(point[5]))
            af_kappa.append(float(point[6]))
    return T, C, M, E, kappa, af_kappa, afm

def plot(points:list,temp:list, name:str):
    y_points = np.array(points)
    x_points = np.array(temp)

    plt.plot(x_points, y_points, label=name, linewidth=1)
    plt.title(name)
    plt.savefig(name+'.png', dpi=300, bbox_inches='tight')
    plt.show()

def main():
    try:
        file_name = sys.argv[1]
    except IndexError:
        print("Указано недостаточное кол-во файлов, укажите файл для ввода данных")
        return
    T, C, M, E, kappa, af_kappa, afm = read_file(file_name)
    plot(C, T, "C")
    plot(M, T, "M")
    plot(afm, T, "Afm")
    plot(E, T, "E")
    plot(kappa, T, "kappa")
    plot(af_kappa, T, "af_kappa")


main()