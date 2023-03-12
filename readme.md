# KY-040 Rotary Encoder Library using GPIOd

Notifies of clockwise and counterclockwise rotation of a KY-040 rotary encoder via a channel.

## Details
Using [gpiod](https://github.com/warthog618/gpiod), an interupt handler is setup to check the line values and push a value to the supplied channel.