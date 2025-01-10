import matplotlib.pyplot as plt
import json
import os
import argparse

MAX_WAVE_LEN = 100000

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

def draw_2waves_same_sample_rate_multichannel(plots, wave1, wave2, ch_amt, opts: Plot_options):
    MAX_LEN = MAX_WAVE_LEN
    if ch_amt == 1:
        MAX_LEN = 10000
    for i in range(ch_amt):
        wave1_cut, wave2_cut = wave1[i::ch_amt][:min(len(wave1), MAX_LEN)], wave2[i::ch_amt][:min(len(wave2), MAX_LEN)]
        draw_2waves_same_sample_rate(plots[i], wave1_cut, wave2_cut, opts.set_title(opts.title + " ch {}".format(i)))
        if opts.with_error and ch_amt == 1:
            err = []
            xs = []
            err_large, xs_large = [], []
            for i in range(min(len(wave1_cut), len(wave2_cut))):
                diff = wave1_cut[i]-wave2_cut[i]
                if abs(diff)>10000: # if difference is too large than it should be easy to find by eyes on other plots but not sure
                    err_large.append(0)
                    xs_large.append(i)
                    continue
                err.append(diff)
                xs.append(i)
            err_opts = opts.set_title(opts.title + " signed error").set_plt_settings([{"c":"red", "s":1, "label":"signed error"}])
            draw_1wave_same_sample_rate(plots[1], xs, err, err_opts, 0)
            draw_1wave_same_sample_rate(plots[1], xs_large, err_large, err_opts.set_plt_settings([{"c":"purple", "s":100, "label":"abs error > 10000"}]), 0) # too large errors

def gcd(a: int, b: int)->int:
    if a < b:
        a, b = b, a
    while b > 0:
        a %= b
        a, b = b, a
    return a

# theoretically its possible to use frac x, but int sol is fine too
def draw_1wave_different_sample_rate(plot, wave1, sr1, sr2, opts: Plot_options, opts_ind: int):
    # r_lcm = (sr1 * sr2) // gcd(sr1, sr2) # not to calc float time - but no point in it
    # mult1, mult2 = r_lcm//sr1, r_lcm//sr2
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

def plot_test_res(fname):
    print("\nWorking on", fname, "\n")
    with open(fname) as f:
        data = json.load(f)
        ch_amt = data["NumChannels"]
        corr_w = data["CorrectW"]
        resampled_w = data["Resampeled"]
        input_w = data["InWave"]
        in_rate = data["InRate"]
        out_rate = data["OutRate"]

        if corr_w != None: # try to draw high res plot with output streams
            fig, ax = plt.subplots(1, 1, figsize=(30, 5))
            opts = Plot_options(title="result wave", plt_settings=[{"c":"red", "s":0.01, "label":"correct wave"}, {"c":"blue", "s":0.01, "label": "resampled wave"}])
            draw_2waves_same_sample_rate_multichannel([ax for i in range(ch_amt)], corr_w, resampled_w, ch_amt, opts)
            fig.savefig(fname + '.svg', dpi=1200)

        fig, ax = plt.subplots(2, 2, figsize=(60, 25))
        opts = Plot_options(plt_settings=[{"c":"red", "s":1, "label": "0 channel"}, {"c":"blue", "s":1, "label": "1 channel"}])
        draw_1wave_different_sample_rate_multichannel(ax[0, 1], input_w, ch_amt, in_rate, out_rate, opts.set_title("input wave"))
        draw_1wave_different_sample_rate_multichannel(ax[1, 1], resampled_w, ch_amt, out_rate, in_rate, opts.set_title("resampled wave"))
        if corr_w != None:
              opts = Plot_options(title="correct wave", plt_settings=[{"c":"red", "s":1, "label":"correct wave"}, {"c":"blue", "s":1, "label": "resampled wave"}], with_error=(True if ch_amt==1 else False))
              draw_2waves_same_sample_rate_multichannel([ax[0, 0], ax[1, 0]], corr_w, resampled_w, ch_amt, opts)
        fig.savefig(fname + '.png', dpi=200)


parser = argparse.ArgumentParser(prog="plots.py", description="plots all saved data from plots/latest")
parser.add_argument("-p", "--plot-path")
plot_path = parser.parse_args().plot_path
for file in os.scandir(plot_path):
    if file.name.endswith(":large"):
        plot_test_res(plot_path + file.name)
