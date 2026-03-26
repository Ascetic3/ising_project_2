#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import json
import os
import sys


OUTPUT_PATH = os.path.join("data", "input", "input.csv")


def mul(a, b):
    params = a.copy()
    params.append(b)
    return params


def export(points, a_steps, m_steps, save):
    with open(OUTPUT_PATH, "a", encoding="utf-8") as output:
        for point in points:
            for parameter in point:
                output.write(f"{parameter};")
            output.write(f"{a_steps};{m_steps};{save}\n")


def cartesian_product(params: list):
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
                res.append(mul(a, b))
        points = res
    return points


def reset():
    os.makedirs(os.path.dirname(OUTPUT_PATH), exist_ok=True)
    with open(OUTPUT_PATH, "w", encoding="utf-8"):
        pass


def fill_parameter_list(p, parameter_list, errors_list=None):
    if errors_list is None:
        errors_list = []
    if isinstance(p, int) or isinstance(p, float):
        parameter_list.append(p)
        return
    if isinstance(p, list):
        for el in p:
            fill_parameter_list(el, parameter_list, errors_list)
        return
    if isinstance(p, dict):
        if "begin" not in p:
            errors_list.append("ошибка диапазона: отсутствует поле begin")
            return
        if "step" not in p:
            errors_list.append("ошибка диапазона: отсутствует поле step")
            return
        if "end" not in p:
            errors_list.append("ошибка диапазона: отсутствует поле end")
            return
        if not (isinstance(p["begin"], int) or isinstance(p["begin"], float)):
            errors_list.append("ошибка диапазона: поле begin должно иметь численное значение")
            return
        if not (isinstance(p["end"], int) or isinstance(p["end"], float)):
            errors_list.append("ошибка диапазона: поле end должно иметь численное значение")
            return
        if not (isinstance(p["step"], int) or isinstance(p["step"], float)):
            errors_list.append("ошибка диапазона: поле step должно иметь численное значение")
            return
        val = begin = p["begin"]
        end = p["end"]
        step = p["step"]

        if begin < end:
            while val < end:
                parameter_list.append(round(val, 6))
                val += abs(step)
                val = round(val, 6)
        if begin > end:
            while val > end:
                parameter_list.append(round(val, 6))
                val -= abs(step)
                val = round(val, 6)
        parameter_list.append(end)
        return


def main():
    reset()

    m_steps_default, a_steps_default = 0, 0
    try:
        file_name = sys.argv[1]
    except IndexError:
        print("Файл задач не найден, укажите файл")
        return

    with open(file_name, encoding="utf-8") as json_file:
        jsn = json.load(json_file)
    tpl = jsn["tpl"]

    for task in jsn["tasks"]:
        save = 0
        data = []
        errors = []
        if "save" in task:
            save = task["save"]
        if "aSteps" in task:
            if not isinstance(task["aSteps"], int):
                print("поле aSteps должно иметь целочисленное значение")
                continue
            a_steps_default = task["aSteps"]
        if "mSteps" in task:
            if not isinstance(task["mSteps"], int):
                print("поле mSteps должно иметь целочисленное значение")
                continue
            m_steps_default = task["mSteps"]
        if a_steps_default + m_steps_default == 0:
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
            export(points, a_steps_default, m_steps_default, save)


if __name__ == "__main__":
    main()
