/*
 * nic_detect.c - Runtime NIC detection and performance recommendations
 *
 * Detects NIC capabilities and recommends optimal configuration:
 * - Identifies the NIC vendor/model from PCI IDs
 * - Reports link speed
 * - Recommends the appropriate backend (AF_XDP, AF_PACKET)
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "reflector.h"

#ifdef __linux__
#include <sys/stat.h>

#include <dirent.h>
#include <unistd.h>
#endif

/* Known NIC vendors, for naming the card and judging whether it is fast
 * enough to be worth warning about when we fall back to AF_PACKET. */
typedef struct {
    uint16_t    vendor_id;
    const char *vendor_name;
    bool        high_perf; /* True if 25G+ capable */
} nic_vendor_t;

static const nic_vendor_t known_nic_vendors[] = {
    {0x8086, "Intel", true},           /* Intel (ixgbe, i40e, ice) */
    {0x15b3, "Mellanox/NVIDIA", true}, /* Mellanox ConnectX */
    {0x14e4, "Broadcom", true},        /* Broadcom NetXtreme */
    {0x1077, "QLogic", false},         /* QLogic */
    {0x177d, "Cavium", true},          /* Cavium ThunderX */
    {0x1d6a, "Aquantia", false},       /* Aquantia */
    {0x1c36, "Amazon ENA", true},      /* AWS ENA */
    {0x1af4, "Virtio", false},         /* Virtio (VMs) */
    {0, NULL, false}                   /* Sentinel */
};

/* Known high-speed NIC models */
typedef struct {
    uint16_t    vendor_id;
    uint16_t    device_id;
    const char *model;
    int         speed_gbps;
} nic_model_t;

static const nic_model_t high_speed_nics[] = {
    /* Intel */
    {0x8086, 0x1572, "Intel X710 (10G)", 10},
    {0x8086, 0x1583, "Intel XL710 (40G)", 40},
    {0x8086, 0x1584, "Intel XXV710 (25G)", 25},
    {0x8086, 0x1592, "Intel E810 (100G)", 100},
    {0x8086, 0x159B, "Intel E810 (25G)", 25},

    /* Mellanox/NVIDIA */
    {0x15b3, 0x1013, "Mellanox ConnectX-4 (100G)", 100},
    {0x15b3, 0x1015, "Mellanox ConnectX-4 Lx (25G)", 25},
    {0x15b3, 0x1017, "Mellanox ConnectX-5 (100G)", 100},
    {0x15b3, 0x1019, "Mellanox ConnectX-5 Ex (100G)", 100},
    {0x15b3, 0x101b, "Mellanox ConnectX-6 (200G)", 200},
    {0x15b3, 0x101d, "Mellanox ConnectX-6 Dx (100G)", 100},
    {0x15b3, 0x101f, "Mellanox ConnectX-6 Lx (25G)", 25},
    {0x15b3, 0x1021, "Mellanox ConnectX-7 (400G)", 400},

    /* Broadcom */
    {0x14e4, 0x16d7, "Broadcom BCM57414 (25G)", 25},
    {0x14e4, 0x16d8, "Broadcom BCM57416 (10G)", 10},

    {0, 0, NULL, 0} /* Sentinel */
};

#ifdef __linux__
/*
 * Read sysfs file and return content as string
 */
static int read_sysfs_str(const char *path, char *buf, size_t buflen)
{
    FILE *f = fopen(path, "r");
    if (!f) {
        return -1;
    }

    if (fgets(buf, buflen, f) == NULL) {
        fclose(f);
        return -1;
    }

    fclose(f);

    /* Remove trailing newline */
    size_t len = strlen(buf);
    if (len > 0 && buf[len - 1] == '\n') {
        buf[len - 1] = '\0';
    }

    return 0;
}

/*
 * Read sysfs file and return as hex value
 */
static int read_sysfs_hex(const char *path, uint16_t *value)
{
    char buf[32];
    if (read_sysfs_str(path, buf, sizeof(buf)) < 0) {
        return -1;
    }

    *value = (uint16_t)strtol(buf, NULL, 16);
    return 0;
}
#endif

/*
 * Get NIC vendor ID from sysfs
 */
int get_nic_vendor(const char *ifname, uint16_t *vendor_id, uint16_t *device_id)
{
#ifdef __linux__
    char path[256];

    snprintf(path, sizeof(path), "/sys/class/net/%s/device/vendor", ifname);
    if (read_sysfs_hex(path, vendor_id) < 0) {
        return -1;
    }

    snprintf(path, sizeof(path), "/sys/class/net/%s/device/device", ifname);
    if (read_sysfs_hex(path, device_id) < 0) {
        *device_id = 0;
    }

    return 0;
#else
    (void)ifname;
    (void)vendor_id;
    (void)device_id;
    return -1; /* Not supported on this platform */
#endif
}

/*
 * Get NIC speed in Mbps from sysfs
 */
int get_nic_speed(const char *ifname)
{
#ifdef __linux__
    char path[256];
    char buf[32];

    snprintf(path, sizeof(path), "/sys/class/net/%s/speed", ifname);
    if (read_sysfs_str(path, buf, sizeof(buf)) < 0) {
        return -1;
    }

    int speed = atoi(buf);
    return speed > 0 ? speed : -1;
#else
    (void)ifname;
    return -1;
#endif
}

/*
 * Detect NIC capabilities and print recommendations
 */
void print_nic_recommendations(const char *ifname)
{
    uint16_t vendor_id = 0;
    uint16_t device_id = 0;
    int      nic_speed = get_nic_speed(ifname);

    /* Try to get NIC vendor info */
    if (get_nic_vendor(ifname, &vendor_id, &device_id) == 0) {
        /* Look up vendor */
        const nic_vendor_t *vendor = NULL;
        for (int i = 0; known_nic_vendors[i].vendor_name; i++) {
            if (known_nic_vendors[i].vendor_id == vendor_id) {
                vendor = &known_nic_vendors[i];
                break;
            }
        }

        /* Look up specific model */
        const nic_model_t *model = NULL;
        for (int i = 0; high_speed_nics[i].model; i++) {
            if (high_speed_nics[i].vendor_id == vendor_id &&
                high_speed_nics[i].device_id == device_id) {
                model = &high_speed_nics[i];
                break;
            }
        }

        /* Print detected NIC info */
        if (model) {
            reflector_log(LOG_INFO, "Detected NIC: %s", model->model);
        } else if (vendor) {
            reflector_log(LOG_INFO, "Detected NIC: %s (ID: %04x:%04x)", vendor->vendor_name,
                          vendor_id, device_id);
        }

        /* Print speed if available */
        if (nic_speed > 0) {
            if (nic_speed >= 1000) {
                reflector_log(LOG_INFO, "Link speed: %d Gbps", nic_speed / 1000);
            } else {
                reflector_log(LOG_INFO, "Link speed: %d Mbps", nic_speed);
            }
        }

        /* Recommendations based on NIC and speed */
        if (nic_speed >= 25000 && vendor && vendor->high_perf) {
            /* AF_XDP is the fastest backend this build has. Say what the
             * ceiling is rather than implying the card's full rate. */
            reflector_log(LOG_INFO, "Using AF_XDP; expect ~40 Gbps, below this NIC's line rate");
        } else if (nic_speed >= 10000) {
            /* 10G NIC - AF_XDP is fine */
            reflector_log(LOG_INFO, "Using AF_XDP (optimal for 10-40G)");
        }

    } else {
        /* Couldn't detect NIC - might be macOS or virtual */
        if (nic_speed > 0) {
            reflector_log(LOG_INFO, "Interface %s: %d Mbps", ifname,
                          nic_speed >= 1000 ? nic_speed / 1000 : nic_speed);
        }
    }

    /* Platform-specific notes */
#ifdef __APPLE__
    reflector_log(LOG_INFO, "Platform: macOS BPF (suitable for development/testing)");
    if (nic_speed > 1000) {
        reflector_log(LOG_INFO, "Tip: For production 10G+ speeds, use a Linux server with AF_XDP");
    }
#endif
}

/*
 * Print AF_PACKET fallback warning with explanation and recommendations
 *
 * Called when AF_XDP is not available and we fall back to AF_PACKET.
 * Explains limitations and suggests how to enable better performance.
 */
void print_af_packet_warning(const char *ifname)
{
    uint16_t vendor_id = 0;
    uint16_t device_id = 0;
    int      nic_speed = get_nic_speed(ifname);

    reflector_log(LOG_WARN, "");
    reflector_log(LOG_WARN, "╔════════════════════════════════════════════════════════════╗");
    reflector_log(LOG_WARN, "║           RUNNING IN AF_PACKET MODE (LIMITED)              ║");
    reflector_log(LOG_WARN, "╠════════════════════════════════════════════════════════════╣");
    reflector_log(LOG_WARN, "║ AF_PACKET is a kernel copy path - expect ~100-500 Mbps max ║");
    reflector_log(LOG_WARN, "║                                                            ║");
    reflector_log(LOG_WARN, "║ WHY: AF_XDP headers not found during build.                ║");
    reflector_log(LOG_WARN, "║      Your kernel may be too old, or libraries missing.     ║");
    reflector_log(LOG_WARN, "║                                                            ║");
    reflector_log(LOG_WARN, "║ TO FIX: Install AF_XDP support:                            ║");
    reflector_log(LOG_WARN, "║   Ubuntu/Debian: sudo apt install libxdp-dev libbpf-dev    ║");
    reflector_log(LOG_WARN, "║   RHEL/Fedora:   sudo dnf install libxdp-devel libbpf-devel║");
    reflector_log(LOG_WARN, "║                                                            ║");
    reflector_log(LOG_WARN, "║ Then rebuild: make clean && make                           ║");
    reflector_log(LOG_WARN, "╚════════════════════════════════════════════════════════════╝");
    reflector_log(LOG_WARN, "");

    /* If they have a fast NIC, this is especially important */
    if (nic_speed >= 10000) {
        reflector_log(LOG_WARN,
                      "Your NIC supports %d Gbps but AF_PACKET will bottleneck at ~500 Mbps!",
                      nic_speed / 1000);
    }

    /* Check NIC and give specific recommendations */
    if (get_nic_vendor(ifname, &vendor_id, &device_id) == 0) {
        const nic_vendor_t *vendor = NULL;
        for (int i = 0; known_nic_vendors[i].vendor_name; i++) {
            if (known_nic_vendors[i].vendor_id == vendor_id) {
                vendor = &known_nic_vendors[i];
                break;
            }
        }

        if (vendor && vendor->high_perf && nic_speed >= 25000) {
            reflector_log(LOG_WARN, "");
            reflector_log(LOG_WARN, "Your %s NIC runs at %dG; AF_XDP is the fastest backend here.",
                          vendor->vendor_name, nic_speed / 1000);
        }
    }
}

/*
 * Print recommended NICs for high-performance use cases
 */
void print_recommended_nics(void)
{
    reflector_log(LOG_INFO, "");
    reflector_log(LOG_INFO, "=== RECOMMENDED NICs FOR HIGH PERFORMANCE ===");
    reflector_log(LOG_INFO, "");
    reflector_log(LOG_INFO, "For AF_XDP (10-40 Gbps, zero-copy):");
    reflector_log(LOG_INFO, "  - Intel X710/XL710 (10G/40G) - Excellent XDP support");
    reflector_log(LOG_INFO, "  - Intel E810 (25G/100G) - Best Intel XDP performance");
    reflector_log(LOG_INFO, "  - Mellanox ConnectX-5/6 (25G-200G) - Native XDP");
    reflector_log(LOG_INFO, "");
    reflector_log(LOG_INFO, "Avoid for high performance:");
    reflector_log(LOG_INFO, "  - Realtek NICs (no XDP support)");
    reflector_log(LOG_INFO, "  - USB NICs (kernel bottleneck)");
    reflector_log(LOG_INFO, "  - Older Intel 1G NICs (e1000, no XDP)");
    reflector_log(LOG_INFO, "");
}
