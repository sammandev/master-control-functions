ALTER TABLE devices
ADD COLUMN source TEXT CHECK (source IN ('BUILT-IN', 'DLL'));
