```
ffmpeg -i d1673d8d-db81-4879-b889-940cd4371735.ivf -c:v h264 -i 206f95e1-bc3e-46c0-b6c1-03f66de41faf.ogg -c:a aac -s 1024x576 test.mp4
ffmpeg -i d1673d8d-db81-4879-b889-940cd4371735.ivf -c:v h264 -i 206f95e1-bc3e-46c0-b6c1-03f66de41faf.ogg -c:a aac -threads 5 -s 1024x576 test2.mp4
ffmpeg.exe -i .\5a6d0592-0271-4f5b-8cdf-7bc1d7fc865e.ogg out.ogg

ffmpeg -i %s -c:v h264 -i %s -c:a aac -s %dx%d %s.mp4
ffmpeg -i %s -c:v h264 -i %s -c:a aac -s %dx%d %s.mp4
ffmpeg.exe -i .\5a6d0592-0271-4f5b-8cdf-7bc1d7fc865e.ogg -threads 5 out.ogg

ffmpeg -i 128b230e-5f14-4c6f-94f3-638a246015db.ivf -c:v h264 -i 39602c61-db0f-40eb-a53d-b006c8babf8d.ogg -c:a aac -s 1024*576 -vsync 2 test3.mp4
```