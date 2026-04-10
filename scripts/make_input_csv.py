#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import sys
import json
import os

# Go читает data/input/input.csv (см. cmd/run).
_INPUT = os.path.join("data", "input", "input.csv")

def multiply(a, b):
    params = a.copy()
    params.append(b)
    pt = params
    return pt
        
def generate_input_csv(points:list, a_steps:int, m_steps:int, save:int):
    with open(_INPUT, "a") as output:
        for point in points:
            complete_point = point
            complete_point.extend([a_steps, m_steps, save])
            output.write(";".join(map(str, complete_point))+"\n")

def cartesian_product(params:list):
    points = []
    for bank in params:
        res = []
        if len(points) == 0:
            for b in bank:
                res.append([b])
            points = res
            continue
        for a in points:
            for b in bank:
                res.append(multiply(a,b))
        points = res
    return points

def clear_input_csv():
    os.makedirs(os.path.dirname(_INPUT), exist_ok=True)
    with open(_INPUT, "w") as f:
        pass

def fill_parameter_list(p, parameterList:list, errorsList = []):
    if isinstance(p, (int, float)):
        parameterList.append(p)
        return
    if isinstance(p, list):
        for el in p:
            fill_parameter_list(el, parameterList, errorsList)
        return
    if isinstance(p, dict):
        #проверки
        for i in ["begin", "step", "end"]:
            if not i in p:
                errorsList.append(f"Ошибка диапазона: отсутствует поле {i}")
                return
            if not (isinstance(p[i], (int, float))):
                errorsList.append(f"Ошибка диапазона: поле {i} должно иметь численное значение")
                return
            
        val = begin = p["begin"]
        end = p["end"]
        step = p["step"]

        if begin < end:
            while val < end:
                parameterList.append(round(val, 6))
                val += abs(step)
                val = round(val, 6)
        if begin > end:
            while val > end:
                parameterList.append(round(val, 6))
                val -= abs(step)
                val = round(val, 6)
        parameterList.append(end)
        return


def main():
    try:
        file_name = sys.argv[1]
    except IndexError:
        print("Файл задач не найден, укажите файл")
        return
    except FileNotFoundError:
        print("Файл указан неверно")
        return
        
    clear_input_csv()

    points = []
    mStepsDefault, aStepsDefault = 0, 0

    with open(file_name) as json_file:
        jsn = json.load(json_file)
    tpl = jsn["tpl"]

    for task in jsn["tasks"]:
        save = 0
        data = []
        errors = []
        if "save" in task:
            save = task["save"]
        if "aSteps" in task:
            if not (isinstance(task["aSteps"], int)):
                print("поле aSteps должно иметь целочисленное значение")
                continue
            aStepsDefault = task["aSteps"]
        if "mSteps" in task:
            if not (isinstance(task["mSteps"], int)):
                print("поле mSteps должно иметь целочисленное значение")
                continue
            mStepsDefault = task["mSteps"]
        if aStepsDefault + mStepsDefault == 0:
            print("Колличество шагов должно быть больше 0")
            continue
        for p in tpl:
            plist = []
            fill_parameter_list(task[p], plist, errors)
            if len(errors) > 0:
                for err in errors:
                    print(err)                   
                    break
            data.append(plist)

        points = cartesian_product(data) 

        if len(errors) == 0:
            generate_input_csv(points, aStepsDefault, mStepsDefault, save)

if __name__ == "__main__":
    main()
