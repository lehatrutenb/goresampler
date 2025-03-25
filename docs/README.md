# go_resampler

Library for sound resampling

### To run test use:
```bash
make clearTestDir
make runTest
```
!CARE make runTest may use lots of RAM - you may try to use make runTestSlow

### To run Benchmark use:
```bash
make clearTestDir
make runBenchmark
```

### To run Benchmark for your own wave use:
```bash
make clearTestDir
make addBaseWave # enter absolute path to sf of wave
make runBenchmarkCustomWave
```

## Resample results
|       /        |                              CONST EXPRESSION RESAMPLER                              |                                   SPLINE RESAMPLER                                   |                                    FFT RESAMPLER                                     |                                  FFMPEG RESAMPLING                                   |
|----------------|--------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| 11025 to 8000  | <video src=https://github.com/user-attachments/assets/839db8f3-953f-4976-ae75-7d37943c54f8> </video> | <video src=https://github.com/user-attachments/assets/1a556220-5397-4189-b026-3db66a148d64> </video> | <video src=https://github.com/user-attachments/assets/bcc2761b-1d3c-4202-86c5-9552b4fd9aa8> </video> | <video src=https://github.com/user-attachments/assets/54eedafa-64c6-4c8d-904a-e041646968d4> </video> |
| 16000 to 8000  | <video src=https://github.com/user-attachments/assets/b3950009-2f16-4348-843d-ad84ebab8ad1> </video> | <video src=https://github.com/user-attachments/assets/748d25df-05e1-47d1-a392-f4ac2fdc57ce> </video> | <video src=https://github.com/user-attachments/assets/d7678eaa-b96b-4a16-86de-c98809457000> </video> | <video src=https://github.com/user-attachments/assets/54eedafa-64c6-4c8d-904a-e041646968d4> </video> |
| 44100 to 8000  | <video src=https://github.com/user-attachments/assets/aa2c55be-bca4-4bc3-8025-286cf0c148c2> </video> | <video src=https://github.com/user-attachments/assets/ccfe1d11-d457-49c1-94e2-537a8550072f> </video> | <video src=https://github.com/user-attachments/assets/06b6f344-13fc-4957-a975-71d631dc85c9> </video> | <video src=https://github.com/user-attachments/assets/54eedafa-64c6-4c8d-904a-e041646968d4> </video> |
| 48000 to 8000  | <video src=https://github.com/user-attachments/assets/c41feee5-85eb-497e-bf9c-d8610a36453c> </video> | <video src=https://github.com/user-attachments/assets/8cbe3a51-1beb-42be-9ba9-fee76c1a634b> </video> | <video src=https://github.com/user-attachments/assets/0644ccf2-4799-40df-bab7-5ec4ace18ca2> </video> | <video src=https://github.com/user-attachments/assets/54eedafa-64c6-4c8d-904a-e041646968d4> </video> |
| 8000 to 16000  | <video src=https://github.com/user-attachments/assets/3d0680f4-f8c0-48c0-8633-85e2eaebbb47> </video> | <video src=https://github.com/user-attachments/assets/e8eea339-0760-4574-9284-fe7fb3a05c29> </video> |                                                                                      | <video src=https://github.com/user-attachments/assets/8fda7b74-193e-4b58-bbd7-71e3f3ec98a4> </video> |
| 11025 to 16000 | <video src=https://github.com/user-attachments/assets/37ab6e0c-dc04-427d-a19c-b573ae14d70c> </video> | <video src=https://github.com/user-attachments/assets/942a3c9a-54f2-459e-8a27-39d7b6382b89> </video> |                                                                                      | <video src=https://github.com/user-attachments/assets/8fda7b74-193e-4b58-bbd7-71e3f3ec98a4> </video> |
| 44100 to 16000 | <video src=https://github.com/user-attachments/assets/436e8aad-e82e-4bf7-8726-8c63e81fd4a1> </video> | <video src=https://github.com/user-attachments/assets/8f8c10f9-cf4c-4f04-8a7b-4537f52a318a> </video> | <video src=https://github.com/user-attachments/assets/74e9461c-83d8-4c9c-81f0-c3fe05713ac6> </video> | <video src=https://github.com/user-attachments/assets/8fda7b74-193e-4b58-bbd7-71e3f3ec98a4> </video> |
| 48000 to 16000 | <video src=https://github.com/user-attachments/assets/8c8965c8-1dab-44bb-8476-0060528900a5> </video> | <video src=https://github.com/user-attachments/assets/ed331651-5f81-49f2-86ba-72f7ca885ac7> </video> | <video src=https://github.com/user-attachments/assets/a45b398b-cabb-4eec-876f-555e83e85a8a> </video> | <video src=https://github.com/user-attachments/assets/8fda7b74-193e-4b58-bbd7-71e3f3ec98a4> </video> |


*** Care, CONSTEXPR RSM in convertations from 11025 to 8000/16000, from 44100 to 8000/16000 rounds 11025 and 44100 to 11000 and 44000