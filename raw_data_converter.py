#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import sys
import os
class Point:
    def C(self):
        C = abs(self.e_sq - (self.e*self.e))/((self.t*self.t)) /(self.L*self.L)
        return C
    
    def kappa(self):
        kappa = abs(self.m_sq - (self.m*self.m))/(self.t)/(self.L*self.L)
        return kappa

    def af_kappa(self):
        af_kappa = abs(self.afm_sq - (self.afm*self.afm))/(self.t)/(self.L*self.L)
        return af_kappa

    def __init__(self, params:list):
        self.L = float(params[0])
        self.t = float(params[1])
        self.e = float(params[2])
        self.e_sq = float(params[3])
        self.m = float(params[4])
        self.m_sq = float(params[5])
        self.afm = float(params[6])
        self.afm_sq = float(params[7])


def read_file(file_name:str):
    points = []
    with open(file_name, "r") as f:
        for line in f:
            point = line.rstrip().split(";")
            points.append(Point(params=[point[0], point[9], *point[13:19]]))
    return points

def export_to_file(file_name:str, points):
    with open(file_name, "w") as f:
        for point in points:
            f.write(f"{point.t};{point.e/(point.L*point.L)};{point.m/(point.L*point.L)};{point.afm/(point.L*point.L)};{point.C()};{point.kappa()};{point.af_kappa()}\n")
    return


def main():
    try:
        import_file = sys.argv[1]
        export_file = sys.argv[2]
    except IndexError:
        print("Указано недостаточное кол-во файлов, укажите файлы для ввода и вывода данных")
        return
    try:
        points = read_file(import_file)
    except FileNotFoundError:
        print("Файл не найден, укажите верный файл")
        return
    try:    
        export_to_file(export_file, points)
    except FileNotFoundError:
        print("Файл не найден, укажите верный файл")
        return

main()