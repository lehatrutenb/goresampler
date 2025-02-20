import matplotlib.pyplot as plt
from multiprocessing import Pool
import json
import os
import argparse

MAX_WAVE_LEN = int(6e4)

class Plot_options:
    def __init__(self, title="", plt_settings=[{}, {}], with_error=False):
        self.title = title
        self.plt_settings = plt_settings
        self.with_error=with_error
    def set_title(self, new_title):
        return Plot_options(title=new_title, plt_settings=self.plt_settings, with_error=self.with_error)
    def set_plt_settings(self, new_plt_settings):
        return Plot_options(plt_settings=new_plt_settings, title=self.title, with_error=self.with_error)
    def set_with_error(self, new_with_error):
        return Plot_options(with_error=new_with_error, title=self.title, plt_settings=self.plt_settings)

def draw_1wave_same_sample_rate(plot, xs, wave, opts: Plot_options, opts_ind: int):
    plot.scatter(xs, wave, **(opts.plt_settings[opts_ind]))
    plot.grid(True)
    plot.set_title(opts.title)
    plot.legend()

def draw_2waves_same_sample_rate(plot, wave1, wave2, opts: Plot_options):
    xs = [i for i in range(len(wave1))]
    draw_1wave_same_sample_rate(plot, xs, wave1, opts, 0)
    draw_1wave_same_sample_rate(plot, xs, wave2, opts, 1)

def draw_error_1wave(plot, waveCorr, waveGot, opts: Plot_options):
    err = []
    xs = []
    err_large, xs_large = [], []
    err20, xs_err20 = [], []
    for i in range(min(len(waveCorr), len(waveGot))):
        diff = waveCorr[i]-waveGot[i]
        if abs(diff)>10000: # if difference is too large than it should be easy to find by eyes on other plots but not sure
            err_large.append(0)
            xs_large.append(i)
            continue
        if abs(diff)>waveCorr[i]*0.2: # if difference is too large than it should be easy to find by eyes on other plots but not sure
            err20.append(diff)
            xs_err20.append(i)
            continue

        err.append(diff)
        xs.append(i)
    err_opts = opts.set_title(opts.title + " signed error").set_plt_settings([{"c":"red", "s":1, "label":"signed error"}])
    draw_1wave_same_sample_rate(plot, xs_err20, err20, err_opts.set_plt_settings([{"c":"teal", "s":5, "label":"abs error > 20%"}]), 0)
    draw_1wave_same_sample_rate(plot, xs, err, err_opts, 0)
    draw_1wave_same_sample_rate(plot, xs_large, err_large, err_opts.set_plt_settings([{"c":"purple", "s":100, "label":"abs error > 10000"}]), 0) # too large errors

def draw_2waves_same_sample_rate_multichannel(plots, wave1, wave2, ch_amt, opts: Plot_options):
    MAX_LEN = MAX_WAVE_LEN
    if ch_amt == 1:
        MAX_LEN = 10000
    for i in range(ch_amt):
        wave1_cut, wave2_cut = wave1[i::ch_amt][:min(len(wave1), MAX_LEN)], wave2[i::ch_amt][:min(len(wave2), MAX_LEN)]
        draw_2waves_same_sample_rate(plots[i], wave1_cut, wave2_cut, opts.set_title(opts.title + " ch {}".format(i)))
        if opts.with_error and ch_amt == 1:
            draw_error_1wave(plots[1], wave1_cut, wave2_cut, opts)
        if opts.with_error and ch_amt == 2:
            draw_error_1wave(plots[i + 2], wave1_cut, wave2_cut, opts)

def gcd(a: int, b: int)->int:
    if a < b:
        a, b = b, a
    while b > 0:
        a %= b
        a, b = b, a
    return a

# theoretically its possible to use frac x, but int sol is fine too
def draw_1wave_different_sample_rate(plot, wave1, sr1, sr2, opts: Plot_options, opts_ind: int):
    st1, st2 = 1/sr1, 1/sr2
    xs = [i*st1 for i in range(len(wave1))]

    plot.scatter(xs, wave1, **(opts.plt_settings[opts_ind]))
    plot.grid(True)
    plot.set_title(opts.title)
    plot.legend()

def draw_1wave_different_sample_rate_multichannel(plot, wave1, ch_amt, sr1, sr2, opts: Plot_options):
    MAX_LEN = MAX_WAVE_LEN # to make waves simular in time
    if sr1 < sr2:
        MAX_LEN = int(float(MAX_LEN) * float(sr1) / float(sr2))
    
    for i in range(ch_amt):
        draw_1wave_different_sample_rate(plot, wave1[i::ch_amt][:min(len(wave1), MAX_LEN)], sr1, sr2, opts, i)

def plot_test_res(f_name_in, f_name_out):
    print("\nWorking on", f_name_in, "\n")
    with open(f_name_in) as f:
        data = json.load(f)
        ch_amt = data["NumChannels"]
        corr_w = data["CorrectW"]
        resampled_w = data["Resampeled"]
        input_w = data["InWave"]
        in_rate = data["InRate"]
        out_rate = data["OutRate"]

        fig, ax = plt.subplots(2 + ch_amt // 2, 2, figsize=(60, 25))
        opts = Plot_options(plt_settings=[{"c":"red", "s":1, "label": "0 channel"}, {"c":"blue", "s":1, "label": "1 channel"}])
        draw_1wave_different_sample_rate_multichannel(ax[0, 1], input_w, ch_amt, in_rate, out_rate, opts.set_title("input wave"))
        draw_1wave_different_sample_rate_multichannel(ax[1, 1], resampled_w, ch_amt, out_rate, in_rate, opts.set_title("resampled wave"))
        if corr_w != None:
            opts = Plot_options(title="correct wave", plt_settings=[{"c":"red", "s":1, "label":"correct wave"}, {"c":"blue", "s":1, "label": "resampled wave"}], with_error=True) # (True if ch_amt==1 else False)
            plots = []
            if ch_amt == 1:
                plots = [ax[0, 0], ax[1, 0]]
            else:
                plots = [ax[0, 0], ax[1, 0], ax[2, 0], ax[2, 1]]

            draw_2waves_same_sample_rate_multichannel(plots, corr_w, resampled_w, ch_amt, opts)
        fig.savefig(f_name_out + '.png', dpi=200)
        plt.close(fig)


parser = argparse.ArgumentParser(prog="plots.py", description="plots all saved data from plots/latest")
parser.add_argument("-j", "--workers-amt")
parser.add_argument("-pib", "--plot-input-base")
parser.add_argument("-pob", "--plot-output-base")
parser.add_argument("-p1", "--plot-path1")
parser.add_argument("-p2", "--plot-path2")
parser.add_argument("-p3", "--plot-path3")
parser.add_argument("-p4", "--plot-path4")
parser.add_argument("-p5", "--plot-path5")
plot_pathes = [parser.parse_args().plot_path1, parser.parse_args().plot_path2, parser.parse_args().plot_path3, parser.parse_args().plot_path4, parser.parse_args().plot_path5]
inBasePath = parser.parse_args().plot_input_base
outBasePath = parser.parse_args().plot_output_base

args = []

for plot_path in plot_pathes:
    cur_in_path = inBasePath+"/"+plot_path+"/"
    cur_out_path = outBasePath+"/"+plot_path+"/"
    for file in os.scandir(cur_in_path):
        if file.name.endswith(":large"):
            args += [(cur_in_path+"/"+file.name, cur_out_path+"/"+file.name)]


p = Pool(int(parser.parse_args().workers_amt))
with p:
    p.starmap(plot_test_res, args)
