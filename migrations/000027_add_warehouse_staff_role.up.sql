ALTER TABLE public.warehouse_staffs
    ADD COLUMN role text NOT NULL DEFAULT 'picker';

ALTER TABLE public.warehouse_staffs
    ADD CONSTRAINT chk_warehouse_staffs_role
    CHECK (role IN ('warehouse_manager','picker','packer','inventory_staff','supervisor'));
