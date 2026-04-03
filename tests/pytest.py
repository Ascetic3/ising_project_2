import matplotlib.pyplot as plt
import numpy as np
import sys
import os
import shutil
import subprocess
from pathlib import Path


def copying(dirname:str, test:str):
    if not os.path.exists(dirname+"/data/input"+test):
        shutil.copyfile(dirname+"/tests/"+test, dirname+"/data/input/input.csv")
    return

def deviation_output(numbers:list):
    with open(dirname+"/tests/deviation.csv", "w") as f:
        for i in range(len(numbers)):
            if i%19==0 and i > 0:
                f.write("\n")
            f.write(f"{numbers[i]};")

    return


def compare(subject:str, reference:str):
    accuracy = []
    test_data = []
    ref_data = []
    with open(subject, "r") as file:
        for line in file:
            test_data.append(list(map(float, (line.rstrip("\n").split(";")))))

    with open(reference, "r") as file:
        for line in file:
            ref_data.append(list(map(float, (line.rstrip("\n").split(";")))))

    for i in range(len(test_data)):
        for j in range(len(test_data[i])):
            try:
                accuracy.append(abs(100 - (test_data[i][j]/(ref_data[i][j]/100))))
            except ZeroDivisionError:
                accuracy.append(float(0))
    print(f"Точность данных -{100 - np.mean(accuracy)}%")
    return accuracy


def main():

    global dirname 
    dirname = str(Path(os.path.dirname(__file__)).parent)

    try:
        test = sys.argv[1]
        ref = dirname+"/tests/"+sys.argv[2]
    except IndexError:
        print("Указано недостаточное кол-во файлов, укажите файл для ввода данных")
        return
    
    if not os.path.exists(dirname+"/data/input"):
        os.makedirs(dirname+"/data/input")
        print("Директория /data/input создана")
    else:
        print("Директория /data/input существует")


    copying(dirname, test)
    print(f"Файл {test} успешно скоприрован")
    
    print("Запуск main.go")
    subprocess.run(["go", "run", f"{dirname}/cmd/run/main.go"])

    if os.path.exists(f"{dirname}/data/output/output.csv"):
        results = f"{dirname}/data/output/output.csv"
    else:
        raise FileNotFoundError
        
    res = compare(results, ref)

    deviation_output(res)
    
    


main()
