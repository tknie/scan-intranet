CREATE TABLE public.scanintranet (
        ip varchar NOT NULL,
        hostname varchar NOT NULL,
        scantime timestamp NOT NULL,
        state varchar DEFAULT 'UNKNOWN'::character varying NOT NULL,
        created timestamp NULL,
        updated_at timestamp NULL
);

CREATE OR REPLACE FUNCTION public.update_timestamp_scan()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
BEGIN
       IF (TG_OP = 'INSERT' ) then 
          NEW.created = CURRENT_TIMESTAMP;
       END IF;
       NEW.updated_at = CURRENT_TIMESTAMP;
       RETURN NEW;
END;
$function$
;

-- Permissions

ALTER FUNCTION public.update_timestamp_scan() OWNER TO postgres;
GRANT ALL ON FUNCTION public.update_timestamp_scan() TO public;
GRANT ALL ON FUNCTION public.update_timestamp_scan() TO postgres;


-- Table Triggers

create trigger update_timestamp_scan before
insert
    or
update
    on
    public.scanintranet for each row execute function update_timestamp_scan();

-- Permissions

ALTER TABLE public.scanintranet OWNER TO postgres;


-- DROP FUNCTION public.update_timestamp_scan();

CREATE OR REPLACE FUNCTION public.update_timestamp_scan()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
BEGIN
       IF (TG_OP = 'INSERT' ) then
          NEW.created = CURRENT_TIMESTAMP;
       END IF;
       NEW.updated_at = CURRENT_TIMESTAMP;
       RETURN NEW;
END;
$function$
;

-- Permissions

ALTER FUNCTION public.update_timestamp_scan() OWNER TO postgres;
GRANT ALL ON FUNCTION public.update_timestamp_scan() TO public;
GRANT ALL ON FUNCTION public.update_timestamp_scan() TO postgres;

