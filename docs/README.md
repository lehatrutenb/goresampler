# go_resampler

Pure go library for sound resampling

# Resamplers desctiption
    It is not not expected for lib users to use raw resamplers - best practise is 
        ResampleBatch with ResamplerAuto (with ResamplerBestFitT) inside
        or 
        ResampleBatch2Waves with ResamplerSpline2Waves inside

    All described resamplers may be found as goresampler.ResamplerT.xxx

### ResamplerConstExprT
    Implements resmapling via filters (КИХ-фильтры)

    + fast
    ~ good quality of resampling
    - can't resample from any x to any y rates (precalced filters restrictions)
    - badly tested on resampling not from {8000, 11000, 16000, 44000, 48000} or not to {8000, 16000}

### ResamplerFFtT
    Has resampling via bluestein FFT inside

    + perfect resampling (in theory)
    - slow
    - can't resample from x to y : y > x (can in theory , but no real point in it currently)
    - can't resample from any x to any y rates (but it is just for safe using)
    - badly tested on resampling not from {8000, 11025, 16000, 44100, 48000} or not to {8000, 16000}

### ResamplerSincT
    Has resampling via sinc on sliding window inside

    + good default resampling quality
    + default speed is near to ResamplerConstExprT (but need precalcs)
    + resampling quality may be configured (window size)
    + speed may be configured (window size)
    + may use both versions that precalcs on resampler creation, and that calcs in resampling time
    ~ much more slower than ResamplerFFtT for large windows
    - large acceptable error rates (>default) may spoil resampling quality
    - badly tested on resampling not from {8000, 11025, 16000, 44100, 48000} or not to {8000, 16000}

### ResamplerSplineT
    Implements resmapling via spline interpollation (of cubic spline with deffect=1)

    ~ two times slower than ResamplerConstExprT
    ~ not very good tested on resampling not from {8000, 11025, 16000, 44100, 48000} or not to {8000, 16000}
    - completely not perfect resampling in frequency domain (in theory)
    - can't resample from any x to any y rates (but it is just for safe using)

### ResamplerSpline2Waves
    Has ResamplerSplineT inside, but resamples to 2 rates without building spline twice (~ two times faster)

    + fast
    ~ not very good tested on resampling not from {8000, 11025, 16000, 44100, 48000} or not to {8000, 16000}
    - completely not perfect resampling in frequency domain (in theory)
    - can't resample from any x to any y rates (but it is just for safe using)

### ResamplerBestFitT
    Has ResamplerConstExprT and ResamplerSincT inside (sinc on 11025 -> 8000/16000 and 44100 -> 8000/16000 resampling)

    + fast
    ~ good quality of resampling
    - can't resample from any x to any y rates (but it is just for safe using)
    - badly tested on resampling not from {8000, 11025, 16000, 44100, 48000} or not to {8000, 16000}

### ResamplerBestFitNotSafeT
    Has ResamplerSincT with no blocks for rate conversions inside

    + fast
    ~ good quality of resampling
    ~ not very good tested on resampling not from {8000, 11025, 16000, 44100, 48000} or not to {8000, 16000}
    - completely not perfect resampling in frequency domain (in theory)
    - can't resample from any x to any y rates (but it is just for safe using)

## Before all
    In test/bechmark cases it is expected to have some base waves for tests/... so
    you may get them via

```bash
make downloadBaseSoundFilesForTests 
```
    Or create your own analog based on structure of mentioned example sound files (using make addBaseWave )

#

### To run tests use:
Output:

./test/plots/ - dir of plots done during testing

./test/audio/ - dir of resampled sound files

./test/!testRes - merged test output

./test/reports/ - dir of reports of resampling with metrics counted during tests

./test/reports/reports_large/ - dir same as reports , but with raw resampling waves too

```bash
make runTest        # runs all internal tests
```
!CARE make runTest may use lots of RAM - you may try to use make runTestSlow
##### Or via act:
```bash
make checkWorkflow   # run same workflow as will be runned on mr
```

#

### To run Benchmark use:
```bash
make clearTestDir   # initialize dir tree for output
make runBenchmark   # runs benchmark ; results are ./test/profile5e5Samples.pdf (profiling of resamplers) ; ./test/readme_audio/listenable/ - resampled audio files
```

### To run Benchmark & get resampled audio for your own wave use:
```bash
make clearTestDir             # initialize dir tree for output
make addBaseWave              # enter absolute path to sound file of yours wave - it will generate necessary sfs for benchmarking
make runBenchmarkCustomWave   # runs benchmark on wave & creates it's resampling results in ./test/readme_audio/listenable/
```

#

## Resample results
|       /        |                              CONST EXPRESSION RESAMPLER                              |                                   SPLINE RESAMPLER                                   |                                    SINC RESAMPLER                                    |                                    FFT RESAMPLER                                     |                                  FFMPEG RESAMPLING                                   |
|----------------|--------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| 11025 to 8000  | <video src=https://github.com/user-attachments/assets/b5e29980-8e64-4c2c-baa9-1f61f30700db> </video> | <video src=https://github.com/user-attachments/assets/894cc88a-2761-4a09-9910-2ada08186ecd> </video> | <video src=https://github.com/user-attachments/assets/008727e3-bf9a-4a81-baeb-0d23aadb2dce> </video> | <video src=https://github.com/user-attachments/assets/573ed728-11ee-4e76-ae67-b4a7182b4a83> </video> | <video src=https://github.com/user-attachments/assets/28d1a848-9066-48b7-b829-bb5260d10319> </video> |
| 16000 to 8000  | <video src=https://github.com/user-attachments/assets/ed22becb-7f4d-49d3-b0f5-14ab84a9708e> </video> | <video src=https://github.com/user-attachments/assets/c902459b-a904-4077-ba1f-29001d721041> </video> | <video src=https://github.com/user-attachments/assets/47557dc4-3869-4131-ab7c-f688119fe065> </video> | <video src=https://github.com/user-attachments/assets/7ca19256-5e96-46ca-9963-86d51439c904> </video> | <video src=https://github.com/user-attachments/assets/28d1a848-9066-48b7-b829-bb5260d10319> </video> |
| 44100 to 8000  | <video src=https://github.com/user-attachments/assets/04d391c7-22b4-4f2e-bce3-0d8727dabc22> </video> | <video src=https://github.com/user-attachments/assets/74fb048a-fdfb-4b07-9aca-206c026d85d8> </video> | <video src=https://github.com/user-attachments/assets/f136b281-69b3-438a-bf10-9546346a37c4> </video> | <video src=https://github.com/user-attachments/assets/ccf51d3d-c834-4c8b-8f7e-62a328dec569> </video> | <video src=https://github.com/user-attachments/assets/28d1a848-9066-48b7-b829-bb5260d10319> </video> |
| 48000 to 8000  | <video src=https://github.com/user-attachments/assets/e3034508-041d-4c51-a94a-e6624feeee9a> </video> | <video src=https://github.com/user-attachments/assets/8d752445-48ba-43c5-a7f8-efba0f799e48> </video> | <video src=https://github.com/user-attachments/assets/5a9d5abb-4555-4f9e-8e96-49ee7010d47d> </video> | <video src=https://github.com/user-attachments/assets/35596824-f13a-4480-89a2-530df5981dcf> </video> | <video src=https://github.com/user-attachments/assets/28d1a848-9066-48b7-b829-bb5260d10319> </video> |
| 8000 to 16000  | <video src=https://github.com/user-attachments/assets/42547eaa-a8dd-46eb-a50d-c54fa088b9bb> </video> | <video src=https://github.com/user-attachments/assets/0ccd683a-7517-4f42-8e7d-e131ca82fabf> </video> | <video src=https://github.com/user-attachments/assets/a20b8541-9bc8-4246-b14b-37556c99c2c7> </video> |                                                                                      | <video src=https://github.com/user-attachments/assets/01435ed9-534f-482e-8ffd-b7963818b11a> </video> |
| 11025 to 16000 | <video src=https://github.com/user-attachments/assets/a67aa243-0f91-4d28-add2-9e7fbc133ba3> </video> | <video src=https://github.com/user-attachments/assets/93ad7783-7537-48b9-84b5-06687ef56b52> </video> | <video src=https://github.com/user-attachments/assets/e21f2459-f301-4761-b24c-ef65cc3668b7> </video> |                                                                                      | <video src=https://github.com/user-attachments/assets/01435ed9-534f-482e-8ffd-b7963818b11a> </video> |
| 44100 to 16000 | <video src=https://github.com/user-attachments/assets/cd0b2cee-0c37-406b-9649-ed873a0591b3> </video> | <video src=https://github.com/user-attachments/assets/0ec3943b-32e0-47c3-be05-8665faf4619b> </video> | <video src=https://github.com/user-attachments/assets/d631b61a-1856-424b-ae3d-b5ca8188f826> </video> | <video src=https://github.com/user-attachments/assets/464ef06a-7f6a-43f9-bf9d-3eff68c9648f> </video> | <video src=https://github.com/user-attachments/assets/01435ed9-534f-482e-8ffd-b7963818b11a> </video> |
| 48000 to 16000 | <video src=https://github.com/user-attachments/assets/5c5e2aeb-1bfa-4795-9c66-77c30cd320df> </video> | <video src=https://github.com/user-attachments/assets/7af672f4-9ee7-4866-8462-f0c4b3cb4862> </video> | <video src=https://github.com/user-attachments/assets/80bfe2c4-24d1-4070-b1d6-344c64f82a5a> </video> | <video src=https://github.com/user-attachments/assets/c0bc8825-9290-4e6e-9301-0a8fe24c0ad2> </video> | <video src=https://github.com/user-attachments/assets/01435ed9-534f-482e-8ffd-b7963818b11a> </video> |

*** Care, CONSTEXPR RSM in convertations from 11025 to 8000/16000, from 44100 to 8000/16000 rounds 11025 and 44100 to 11000 and 44000