from scipy import integrate
import numpy as np
import matplotlib.pyplot as plt
import sys

def read_file(file_name:str):
    var1 = []
    var2 = []
    with open(file_name, "r", encoding="utf-8") as f:
        for line in f:
            point = line.rstrip().split(";")
            var1.append(float(point[0]))
            var2.append(float(point[4]))
    return var1, var2


def integration(temp, C):
    temp = np.array(temp)
    C = np.array(C)
    C = C/temp
    return integrate.cumulative_simpson(x=temp, y=C, initial=0)

def make_plot(x, y):
    print(len(x), len(y))
    x_points = np.array(x)
    y_points = np.array(y)
    plt.plot(x_points, y_points, label="Энтропия", linewidth=1)
    plt.title("Энтропия")
    plt.xlabel("T")
    plt.ylabel("Энтропия")
    plt.legend()
    plt.savefig("Энтропия" + '.png', dpi=300, bbox_inches='tight')
    plt.show()


def main():
    try:
        import_file = sys.argv[1]
    except IndexError:
        print('Не указан файл')
        return
    try:
        var1, var2 = read_file(import_file)
    except FileNotFoundError:
        print("Файл не найден, укажите верный файл")
        return
    var1 = var1[::-1]
    var2 = var2[::-1]
    y = integration(var1, var2)
    make_plot(var1, y)

main()
