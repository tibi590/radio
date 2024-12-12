#include "c_main.h"

#include <stdio.h>
#include <unistd.h>
#include <sys/io.h>

#define base 0x378 // Base address of the parallel port

#define BIT(nth_bit)                    (1U << (nth_bit))
#define CHECK_BIT(data, bit)            ((data) & BIT(bit))
#define SET_BIT(data, bit)              ((data) |= BIT(bit))
#define CLEAR_BIT(data, bit)            ((data) &= ~BIT(bit))
#define CHANGE_BIT(data, bit)           ((data) ^= BIT(bit))

bool enable_perm() {
    if (ioperm(base, 1, 1)) {       
        printf("Access denied to %x\n", base);
        return false;
    }
    return true;
}

bool disable_perm() {
    if (ioperm(base, 1, 0)) {
        printf("Couldn't revoke access to %x\n", base);
        return false;
    }
    return true;
}

void set_pins(unsigned char status_bits) {
    outb(status_bits, base);
}