package runtime

func Run(cfg Config) error {
    rootfs, cleanup, err := mountSnapshot(cfg)
    if err != nil {
        return err
    }
    defer cleanup()

    return start(rootfs, cfg)
}