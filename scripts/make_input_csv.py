#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import sys
import json

def mul(a, b):
    params = a.copy()
    params.append(b)
    pt = params
    return pt
        
def export(points:list, a_steps:int, m_steps:int, save:int):
    with open(f"input.csv", "a") as output:
        for point in points:
            for parameter in point:
                output.write(f"{parameter};")
            output.write(f"{a_steps};{m_steps};{save}\n")

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
                res.append(mul(a,b))
        points = res
    return points

def reset():
    with open("input.csv", "w") as f:
        pass

def fillParameterList(p, parameterList, errorsList = []):
    if isinstance(p, int) or isinstance(p, float):
        parameterList.append((p))
        return
    if isinstance(p, list):
        for el in p:
            fillParameterList(el,parameterList,errorsList)
        return
    if isinstance(p, dict):
        #проверки
        if not "begin" in p:
            errorsList.append("ошибка диапазона: отсутствует поле begin")
            return
        if not "step" in p:
            errorsList.append("ошибка диапазона: отсутствует поле step")
            return
        if not "end" in p:
            errorsList.append("ошибка диапазона: отсутствует поле end")
            return
        if not (isinstance(p["begin"], int) or isinstance(p["begin"], float)):
            errorsList.append("ошибка диапазона: поле begin должно иметь численное значение")
            return
        if not (isinstance(p["end"], int) or isinstance(p["end"], float)):
            errorsList.append("ошибка диапазона: поле end должно иметь численное значение")
            return
        if not (isinstance(p["step"], int) or isinstance(p["step"], float)):
            errorsList.append("ошибка диапазона: поле step должно иметь численное значение")
            return
        val = begin = p["begin"]
        end = p["end"]
        step = p["step"]

        if begin < end:
            while val < end:
                parameterList.append((round(val,6)))
                val += abs(step)
                val = round(val,6)
        if begin > end:
            while val > end:
                parameterList.append((round(val,6)))
                val -= abs(step)
                val = round(val,6)
        parameterList.append((end))
        return


def main():
    
    reset()
    points = []
    mStepsDefault, aStepsDefault = 0, 0
    try:
        file_name = sys.argv[1]
    except IndexError:
        print("Файл задач не найден, укажите файл")
        return
    except FileNotFoundError:
        print("Файл указан неверно")
        return

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
            continue
        for p in tpl:
            plist = []
            fillParameterList(task[p], plist, errors)
            if len(errors) > 0:
                for err in errors:
                    print(err)                   
                    break
            data.append(plist)

            points = cartesian_product(data) 
        if len(errors) == 0:
            export(points, aStepsDefault, mStepsDefault, save)

if __name__ == "__main__":
    main()