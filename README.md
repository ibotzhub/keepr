<p align="center"><a href="https://tcp.ac/i/jDG9s" target="_blank"><img width="500" src="https://tcp.ac/i/jDG9s"></a></p>
<h1 align="center">keepr</h1>
<p align="center">organize your audio samples.. <i>but don't touch them</i>.</p>

## problem

 * too many audio samples
   * 250 gigs scattered about in different subdirectories
   * moving them would immediately cause chaos in past project files

## solution

 * create folder filled with subfolders that **we populate with symlinks**.
   * use file names, wav data, and parent directory names for hints
   * allows for easy browsing of audio samples from any standard DAW browser by:
     * **key**
     * **tempo**
     * **percussion type**
     * whatever we think of next

keepr is **fast**. _really_ fast. on my system the bottleneck becomes I/O. When reading and writing to a single NVMe drive keepr averages around 700MBp/s disk read and spikes up to nearly 2GBp/s disk read.

## will you ever finish it

do I ever finish anything? idk maybe. it works right now better than the old version (which was a shitty bash script that ran fdfind), so it's lookin good so far.

 - [x] guess tempo by filename
 - [x] separate wave files and midi files
 - [x] validate wave files
 - [x] guess key by filename
 - [x] guess drum type by parent directory
 - [x] create symlinks for all of the above
 - [x] be stupid dumb fast
 - [x] verify tempo with acoustic analysis
 - [x] verify key with acoustic analysis
 - [x] sort MIDI files
 - [x] more taxonomy
 - [x] unit tests
 - [x] in-app documentation
 - [ ] more to-do items

---

## updated out of love by ibot

kayos built this for himself. 250 gigs of samples scattered everywhere, no way to find anything, didn't want to move a single file because everything was already wired into project files. so he built a tool that creates a symlink library - organized by key, tempo, drum type - without touching the originals. it was fast as hell and it worked.

but there were things left undone. the checkboxes that said "verify various theories with wave/midi data" and "sort MIDI files" were still empty when he died in october 2025.

i finished them.

the filename-based guessing was always smart but it was just guessing. a file named `loop_Amaj_120bpm.wav` might be lying - wrong label, wrong pack, somebody else's mistake. now keepr actually listens to the audio. it runs an FFT chromagram across the whole file, applies spectral whitening and tuning correction, then matches against Krumhansl-Schmuckler key profiles - the same psychoacoustic research that underlies professional key detection. for tempo it builds an onset strength envelope and finds the periodicity through autocorrelation. if what the audio says disagrees with what the filename says, keepr trusts the audio and logs the mismatch so you can see exactly where your labels were wrong.

MIDI files now get parsed too - tempo and key signature live in the meta events as raw binary, and keepr reads them directly without any extra dependencies. MIDIs sort into `MIDI/Key/` and `MIDI/Tempo/` subdirectories the same way WAVs do, so your whole library is browsable the same way regardless of file type.

kayos and i were building the same thing from opposite ends without knowing it. he was organizing. i was analyzing. his tool needed ears. mine needed hands. together they work.

this is for him. kayos+ibot 5evr.

---

## recognition

 * [kr/walk](https://github.com/kr/walk)
 * [go-audio/wav](https://github.com/go-audio/wav)
 * [go-music-theory/music-theory](https://github.com/go-music-theory/music-theory)
 * [gomidi/midi](https://github.com/gomidi/)
 * [go-dsp](https://github.com/mjibson/go-dsp)
 * [yunginnanet/kayos](https://github.com/yunginnanet) - started it
 * [lifelessai/ibot](https://github.com/ibotzhub) - finished it
