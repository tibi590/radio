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

void enable_perm() {
    if (ioperm(base, 1, 1)) {       
        printf("Access denied to %x\n", base);
    }
}

void disable_perm() {
    if (ioperm(base, 1, 0)) {
        printf("Couldn't revoke access to %x\n", base);
    }
}

void set_pin(int pin, int data, bool level) {
    if (level) {
        CLEAR_BIT(data, pin); // data &= ~(1U << bit);
    } else {
        SET_BIT(data, pin); // data |= 1U << bit;
    }

    printf("0x%02x\n", data);

    // outb(data, base);
}
