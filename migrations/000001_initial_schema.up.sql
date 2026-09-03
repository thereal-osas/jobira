-- Jobira baseline schema migration
-- Generated from the PostgreSQL 18.3 schema dump.
-- Canonical favourites table: favorite_cleaners.
-- Obsolete objects intentionally excluded: favourite_cleaners and legacy_messages.

--
-- PostgreSQL database dump
--


-- Dumped from database version 18.3
-- Dumped by pg_dump version 18.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: application_timeline; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.application_timeline (
    id bigint NOT NULL,
    application_id bigint NOT NULL,
    status text NOT NULL,
    actor_user_id bigint,
    note text DEFAULT ''::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT application_timeline_valid_status CHECK ((status = ANY (ARRAY['pending'::text, 'shortlisted'::text, 'invited'::text, 'accepted'::text, 'rejected'::text, 'completed'::text, 'cancelled'::text])))
);

--
-- Name: application_timeline_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.application_timeline_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: application_timeline_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.application_timeline_id_seq OWNED BY public.application_timeline.id;

--
-- Name: applications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.applications (
    id bigint NOT NULL,
    job_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    cover_message text NOT NULL,
    proposed_rate integer DEFAULT 0 NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: applications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.applications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: applications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.applications_id_seq OWNED BY public.applications.id;

--
-- Name: availability_overrides; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.availability_overrides (
    id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    available_date date NOT NULL,
    start_time time without time zone NOT NULL,
    end_time time without time zone NOT NULL,
    status character varying(30) DEFAULT 'unavailable'::character varying NOT NULL,
    reason text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT availability_overrides_check CHECK ((end_time > start_time))
);

--
-- Name: availability_overrides_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.availability_overrides_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: availability_overrides_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.availability_overrides_id_seq OWNED BY public.availability_overrides.id;

--
-- Name: availability_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.availability_settings (
    cleaner_id bigint NOT NULL,
    min_notice_minutes integer DEFAULT 120 NOT NULL,
    min_booking_minutes integer DEFAULT 60 NOT NULL,
    max_booking_minutes integer DEFAULT 480 NOT NULL,
    buffer_minutes integer DEFAULT 30 NOT NULL,
    booking_horizon_days integer DEFAULT 90 NOT NULL,
    timezone character varying(100) DEFAULT 'Europe/London'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: billing_checkout_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.billing_checkout_sessions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    plan_id bigint,
    stripe_session_id text NOT NULL,
    stripe_customer_id text,
    stripe_subscription_id text,
    status text DEFAULT 'created'::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: billing_checkout_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.billing_checkout_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: billing_checkout_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.billing_checkout_sessions_id_seq OWNED BY public.billing_checkout_sessions.id;

--
-- Name: billing_customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.billing_customers (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    stripe_customer_id text NOT NULL,
    email text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: billing_customers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.billing_customers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: billing_customers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.billing_customers_id_seq OWNED BY public.billing_customers.id;

--
-- Name: blocked_cleaners; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blocked_cleaners (
    id bigint NOT NULL,
    client_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    created_at timestamp without time zone NOT NULL,
    reason text DEFAULT ''::text NOT NULL
);

--
-- Name: blocked_cleaners_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.blocked_cleaners_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: blocked_cleaners_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.blocked_cleaners_id_seq OWNED BY public.blocked_cleaners.id;

--
-- Name: booking_status_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.booking_status_history (
    id bigint NOT NULL,
    booking_id bigint NOT NULL,
    changed_by bigint,
    from_status text,
    to_status text NOT NULL,
    note text,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: booking_status_history_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.booking_status_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: booking_status_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.booking_status_history_id_seq OWNED BY public.booking_status_history.id;

--
-- Name: bookings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.bookings (
    id bigint NOT NULL,
    job_id bigint NOT NULL,
    application_id bigint,
    client_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    scheduled_at timestamp without time zone,
    completed_at timestamp without time zone,
    cancelled_at timestamp without time zone,
    cancellation_reason text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    closed_at timestamp without time zone,
    closure_status text,
    closure_comment text,
    client_confirmed_completion boolean DEFAULT false NOT NULL,
    client_would_hire_again boolean,
    client_rating integer,
    scheduled_end_at timestamp with time zone,
    cancelled_by bigint
);

--
-- Name: bookings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.bookings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: bookings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.bookings_id_seq OWNED BY public.bookings.id;

--
-- Name: cleaner_availability; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cleaner_availability (
    id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    available_date date NOT NULL,
    start_time text NOT NULL,
    end_time text NOT NULL,
    status text DEFAULT 'available'::text NOT NULL,
    notes text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: cleaner_availability_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cleaner_availability_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: cleaner_availability_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cleaner_availability_id_seq OWNED BY public.cleaner_availability.id;

--
-- Name: cleaner_available_now; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cleaner_available_now (
    id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    is_available boolean DEFAULT false NOT NULL,
    available_from timestamp without time zone NOT NULL,
    available_until timestamp without time zone NOT NULL,
    location text DEFAULT ''::text NOT NULL,
    travel_radius_miles integer DEFAULT 0 NOT NULL,
    job_types text[] DEFAULT '{}'::text[] NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT cleaner_available_now_valid_radius CHECK ((travel_radius_miles >= 0)),
    CONSTRAINT cleaner_available_now_valid_window CHECK ((available_until > available_from))
);

--
-- Name: cleaner_available_now_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cleaner_available_now_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: cleaner_available_now_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cleaner_available_now_id_seq OWNED BY public.cleaner_available_now.id;

--
-- Name: cleaner_profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cleaner_profiles (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    bio text NOT NULL,
    location text NOT NULL,
    years_experience integer DEFAULT 0 NOT NULL,
    hourly_rate integer DEFAULT 0 NOT NULL,
    services_offered text NOT NULL,
    is_verified boolean DEFAULT false NOT NULL,
    verification_status text DEFAULT 'pending'::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    country text DEFAULT 'UK'::text NOT NULL,
    city text DEFAULT 'London'::text NOT NULL,
    region text DEFAULT 'unknown'::text NOT NULL,
    postcode_area text DEFAULT 'unknown'::text NOT NULL,
    availability_status text DEFAULT 'available_immediately'::text NOT NULL,
    travel_radius_miles integer DEFAULT 10 NOT NULL,
    jobs_completed integer DEFAULT 0 NOT NULL,
    jobs_cancelled integer DEFAULT 0 NOT NULL,
    response_rate integer DEFAULT 100 NOT NULL,
    reliability_score integer DEFAULT 0 NOT NULL,
    badge text DEFAULT 'New Cleaner'::text NOT NULL
);

--
-- Name: cleaner_profiles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cleaner_profiles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: cleaner_profiles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cleaner_profiles_id_seq OWNED BY public.cleaner_profiles.id;

--
-- Name: cleaner_reports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cleaner_reports (
    id bigint NOT NULL,
    client_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    reason text NOT NULL,
    details text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'open'::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    resolved_by bigint,
    resolved_at timestamp without time zone,
    admin_notes text DEFAULT ''::text
);

--
-- Name: cleaner_reports_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cleaner_reports_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: cleaner_reports_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cleaner_reports_id_seq OWNED BY public.cleaner_reports.id;

--
-- Name: cleaner_reputation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cleaner_reputation (
    cleaner_id bigint NOT NULL,
    average_rating numeric(3,2) DEFAULT 0 NOT NULL,
    total_reviews integer DEFAULT 0 NOT NULL,
    completed_jobs integer DEFAULT 0 NOT NULL,
    repeat_clients integer DEFAULT 0 NOT NULL,
    would_hire_again_count integer DEFAULT 0 NOT NULL,
    recommendation_percentage integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    badge text DEFAULT 'New Cleaner'::text NOT NULL,
    total_bookings integer DEFAULT 0 NOT NULL,
    cleaner_cancellations integer DEFAULT 0 NOT NULL,
    eligible_response_messages integer DEFAULT 0 NOT NULL,
    responded_messages integer DEFAULT 0 NOT NULL,
    average_response_minutes integer DEFAULT 0 NOT NULL
);

--
-- Name: client_cleaner_notes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.client_cleaner_notes (
    id bigint NOT NULL,
    client_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    note text NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);

--
-- Name: client_cleaner_notes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.client_cleaner_notes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: client_cleaner_notes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.client_cleaner_notes_id_seq OWNED BY public.client_cleaner_notes.id;

--
-- Name: companies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.companies (
    id bigint NOT NULL,
    owner_id bigint NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: companies_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.companies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: companies_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.companies_id_seq OWNED BY public.companies.id;

--
-- Name: company_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.company_members (
    id bigint NOT NULL,
    company_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    role text DEFAULT 'cleaner'::text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    CONSTRAINT company_members_valid_role CHECK ((role = ANY (ARRAY['owner'::text, 'admin'::text, 'cleaner'::text]))),
    CONSTRAINT company_members_valid_status CHECK ((status = ANY (ARRAY['active'::text, 'inactive'::text])))
);

--
-- Name: company_members_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.company_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: company_members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.company_members_id_seq OWNED BY public.company_members.id;

--
-- Name: conversations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.conversations (
    id bigint NOT NULL,
    booking_id bigint NOT NULL,
    client_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    last_message_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT conversations_different_users_check CHECK ((client_id <> cleaner_id)),
    CONSTRAINT conversations_status_check CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'archived'::character varying, 'blocked'::character varying])::text[])))
);

--
-- Name: conversations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.conversations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: conversations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.conversations_id_seq OWNED BY public.conversations.id;

--
-- Name: favorite_cleaners; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.favorite_cleaners (
    id bigint NOT NULL,
    client_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: favorite_cleaners_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.favorite_cleaners_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: favorite_cleaners_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.favorite_cleaners_id_seq OWNED BY public.favorite_cleaners.id;

--
-- Name: job_alerts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.job_alerts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    location text NOT NULL,
    job_type text NOT NULL,
    minimum_budget integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: job_alerts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.job_alerts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: job_alerts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.job_alerts_id_seq OWNED BY public.job_alerts.id;

--
-- Name: job_invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.job_invitations (
    id bigint NOT NULL,
    job_id bigint NOT NULL,
    client_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    status text DEFAULT 'sent'::text NOT NULL,
    message text DEFAULT ''::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: job_invitations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.job_invitations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: job_invitations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.job_invitations_id_seq OWNED BY public.job_invitations.id;

--
-- Name: jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.jobs (
    id bigint NOT NULL,
    client_id bigint NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    location text NOT NULL,
    job_type text NOT NULL,
    budget integer DEFAULT 0 NOT NULL,
    status text DEFAULT 'open'::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    listing_type text DEFAULT 'shift'::text NOT NULL
);

--
-- Name: jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.jobs_id_seq OWNED BY public.jobs.id;

--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.messages (
    id bigint CONSTRAINT messages_id_not_null1 NOT NULL,
    conversation_id bigint NOT NULL,
    sender_id bigint CONSTRAINT messages_sender_id_not_null1 NOT NULL,
    message_type character varying(30) DEFAULT 'text'::character varying NOT NULL,
    content text CONSTRAINT messages_content_not_null1 NOT NULL,
    is_read boolean DEFAULT false CONSTRAINT messages_is_read_not_null1 NOT NULL,
    read_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() CONSTRAINT messages_created_at_not_null1 NOT NULL,
    updated_at timestamp without time zone DEFAULT now() CONSTRAINT messages_updated_at_not_null1 NOT NULL,
    CONSTRAINT messages_type_check CHECK (((message_type)::text = ANY ((ARRAY['text'::character varying, 'image'::character varying, 'system'::character varying])::text[])))
);

--
-- Name: messages_id_seq1; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.messages_id_seq1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: messages_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.messages_id_seq1 OWNED BY public.messages.id;

--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    title text NOT NULL,
    message text NOT NULL,
    type text NOT NULL,
    is_read boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;

--
-- Name: platform_access_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.platform_access_settings (
    id bigint NOT NULL,
    launch_grace_started_at timestamp without time zone,
    launch_grace_days integer DEFAULT 60 NOT NULL,
    launch_grace_enabled boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: platform_access_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.platform_access_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: platform_access_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.platform_access_settings_id_seq OWNED BY public.platform_access_settings.id;

--
-- Name: preferred_cleaners; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.preferred_cleaners (
    id bigint NOT NULL,
    client_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: preferred_cleaners_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.preferred_cleaners_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: preferred_cleaners_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.preferred_cleaners_id_seq OWNED BY public.preferred_cleaners.id;

--
-- Name: profile_media; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.profile_media (
    id bigint NOT NULL,
    owner_user_id bigint,
    company_id bigint,
    media_type text NOT NULL,
    url text NOT NULL,
    caption text DEFAULT ''::text NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT profile_media_owner_check CHECK (((owner_user_id IS NOT NULL) OR (company_id IS NOT NULL))),
    CONSTRAINT profile_media_valid_type CHECK ((media_type = ANY (ARRAY['profile_photo'::text, 'portfolio'::text, 'pricing'::text, 'company_logo'::text])))
);

--
-- Name: profile_media_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.profile_media_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: profile_media_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.profile_media_id_seq OWNED BY public.profile_media.id;

--
-- Name: promo_codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.promo_codes (
    id bigint NOT NULL,
    code text NOT NULL,
    description text,
    discount_type text NOT NULL,
    percentage_off integer DEFAULT 0 NOT NULL,
    fixed_amount_pence integer DEFAULT 0 NOT NULL,
    free_months integer DEFAULT 0 NOT NULL,
    max_uses integer DEFAULT 0 NOT NULL,
    times_used integer DEFAULT 0 NOT NULL,
    user_type text,
    plan_id bigint,
    starts_at timestamp without time zone,
    expires_at timestamp without time zone,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    stripe_coupon_id text,
    stripe_promotion_code_id text
);

--
-- Name: promo_codes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.promo_codes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: promo_codes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.promo_codes_id_seq OWNED BY public.promo_codes.id;

--
-- Name: recently_viewed_cleaners; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.recently_viewed_cleaners (
    id bigint NOT NULL,
    client_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    viewed_at timestamp without time zone NOT NULL
);

--
-- Name: recently_viewed_cleaners_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.recently_viewed_cleaners_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: recently_viewed_cleaners_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.recently_viewed_cleaners_id_seq OWNED BY public.recently_viewed_cleaners.id;

--
-- Name: recurring_availability; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.recurring_availability (
    id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    weekday integer NOT NULL,
    start_time time without time zone NOT NULL,
    end_time time without time zone NOT NULL,
    status character varying(30) DEFAULT 'available'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT recurring_availability_check CHECK ((end_time > start_time)),
    CONSTRAINT recurring_availability_weekday_check CHECK (((weekday >= 0) AND (weekday <= 6)))
);

--
-- Name: recurring_availability_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.recurring_availability_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: recurring_availability_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.recurring_availability_id_seq OWNED BY public.recurring_availability.id;

--
-- Name: referral_codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.referral_codes (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    code text NOT NULL,
    reward_type text DEFAULT 'credit'::text NOT NULL,
    reward_value_pence integer DEFAULT 0 NOT NULL,
    max_uses integer DEFAULT 0 NOT NULL,
    times_used integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: referral_codes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.referral_codes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: referral_codes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.referral_codes_id_seq OWNED BY public.referral_codes.id;

--
-- Name: referral_redemptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.referral_redemptions (
    id bigint NOT NULL,
    referral_code_id bigint NOT NULL,
    referrer_id bigint NOT NULL,
    referred_user_id bigint NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    reward_applied boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: referral_redemptions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.referral_redemptions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: referral_redemptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.referral_redemptions_id_seq OWNED BY public.referral_redemptions.id;

--
-- Name: repeat_booking_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.repeat_booking_requests (
    id bigint NOT NULL,
    original_booking_id bigint NOT NULL,
    new_booking_id bigint,
    client_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    job_id bigint NOT NULL,
    scheduled_at timestamp without time zone NOT NULL,
    status text DEFAULT 'created'::text NOT NULL,
    message text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    scheduled_end_at timestamp without time zone
);

--
-- Name: repeat_booking_requests_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.repeat_booking_requests_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: repeat_booking_requests_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.repeat_booking_requests_id_seq OWNED BY public.repeat_booking_requests.id;

--
-- Name: repeat_bookings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.repeat_bookings (
    id bigint NOT NULL,
    client_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    job_id bigint NOT NULL,
    message text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'sent'::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: repeat_bookings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.repeat_bookings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: repeat_bookings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.repeat_bookings_id_seq OWNED BY public.repeat_bookings.id;

--
-- Name: reports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reports (
    id bigint NOT NULL,
    reporter_id bigint NOT NULL,
    reported_user_id bigint,
    job_id bigint,
    booking_id bigint,
    report_type text NOT NULL,
    reason text NOT NULL,
    details text,
    status text DEFAULT 'open'::text NOT NULL,
    admin_notes text,
    reviewed_by bigint,
    reviewed_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: reports_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reports_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: reports_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reports_id_seq OWNED BY public.reports.id;

--
-- Name: reviews; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reviews (
    id bigint NOT NULL,
    job_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    client_id bigint NOT NULL,
    rating integer NOT NULL,
    comment text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    booking_id bigint
);

--
-- Name: reviews_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reviews_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: reviews_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reviews_id_seq OWNED BY public.reviews.id;

--
-- Name: saved_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saved_jobs (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    job_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: saved_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saved_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: saved_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saved_jobs_id_seq OWNED BY public.saved_jobs.id;

--
-- Name: subscription_plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscription_plans (
    id bigint NOT NULL,
    name text NOT NULL,
    role_type text NOT NULL,
    price_pence integer DEFAULT 0 NOT NULL,
    billing_interval text DEFAULT 'monthly'::text NOT NULL,
    application_limit integer DEFAULT 5 NOT NULL,
    job_post_limit integer DEFAULT 5 NOT NULL,
    cleaner_seat_limit integer DEFAULT 1 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    stripe_price_id text
);

--
-- Name: subscription_plans_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.subscription_plans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: subscription_plans_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.subscription_plans_id_seq OWNED BY public.subscription_plans.id;

--
-- Name: user_daily_access; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_daily_access (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    access_date date DEFAULT CURRENT_DATE NOT NULL,
    applications_today integer DEFAULT 0 NOT NULL,
    jobs_posted_today integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: user_daily_access_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_daily_access_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: user_daily_access_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_daily_access_id_seq OWNED BY public.user_daily_access.id;

--
-- Name: user_subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_subscriptions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    plan_id bigint,
    status text DEFAULT 'trial'::text NOT NULL,
    trial_started_at timestamp without time zone DEFAULT now() NOT NULL,
    trial_ends_at timestamp without time zone,
    current_period_start timestamp without time zone,
    current_period_end timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    stripe_subscription_id text
);

--
-- Name: user_subscriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_subscriptions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: user_subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_subscriptions_id_seq OWNED BY public.user_subscriptions.id;

--
-- Name: user_usage; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_usage (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    application_count integer DEFAULT 0 NOT NULL,
    job_post_count integer DEFAULT 0 NOT NULL,
    free_application_limit integer DEFAULT 5 NOT NULL,
    free_job_post_limit integer DEFAULT 5 NOT NULL,
    monetisation_enabled boolean DEFAULT false NOT NULL,
    trial_started_at timestamp without time zone DEFAULT now() NOT NULL,
    trial_ends_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    trial_days integer DEFAULT 14 NOT NULL,
    daily_application_limit integer DEFAULT 5 NOT NULL,
    free_job_visibility_delay_minutes integer DEFAULT 20 NOT NULL,
    daily_job_post_limit integer DEFAULT 5 NOT NULL
);

--
-- Name: user_usage_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_usage_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: user_usage_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_usage_id_seq OWNED BY public.user_usage.id;

--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    full_name text NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    role text DEFAULT 'user'::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

--
-- Name: verification_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.verification_requests (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    verification_type text NOT NULL,
    document_url text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    admin_notes text,
    reviewed_by bigint,
    reviewed_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Name: verification_requests_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.verification_requests_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: verification_requests_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.verification_requests_id_seq OWNED BY public.verification_requests.id;

--
-- Name: work_proofs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.work_proofs (
    id bigint NOT NULL,
    booking_id bigint NOT NULL,
    cleaner_id bigint NOT NULL,
    proof_type text NOT NULL,
    photo_url text NOT NULL,
    caption text DEFAULT ''::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT work_proofs_valid_type CHECK ((proof_type = ANY (ARRAY['before'::text, 'after'::text])))
);

--
-- Name: work_proofs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.work_proofs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: work_proofs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.work_proofs_id_seq OWNED BY public.work_proofs.id;

--
-- Name: application_timeline id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.application_timeline ALTER COLUMN id SET DEFAULT nextval('public.application_timeline_id_seq'::regclass);

--
-- Name: applications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applications ALTER COLUMN id SET DEFAULT nextval('public.applications_id_seq'::regclass);

--
-- Name: availability_overrides id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.availability_overrides ALTER COLUMN id SET DEFAULT nextval('public.availability_overrides_id_seq'::regclass);

--
-- Name: billing_checkout_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_checkout_sessions ALTER COLUMN id SET DEFAULT nextval('public.billing_checkout_sessions_id_seq'::regclass);

--
-- Name: billing_customers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_customers ALTER COLUMN id SET DEFAULT nextval('public.billing_customers_id_seq'::regclass);

--
-- Name: blocked_cleaners id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocked_cleaners ALTER COLUMN id SET DEFAULT nextval('public.blocked_cleaners_id_seq'::regclass);

--
-- Name: booking_status_history id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.booking_status_history ALTER COLUMN id SET DEFAULT nextval('public.booking_status_history_id_seq'::regclass);

--
-- Name: bookings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bookings ALTER COLUMN id SET DEFAULT nextval('public.bookings_id_seq'::regclass);

--
-- Name: cleaner_availability id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_availability ALTER COLUMN id SET DEFAULT nextval('public.cleaner_availability_id_seq'::regclass);

--
-- Name: cleaner_available_now id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_available_now ALTER COLUMN id SET DEFAULT nextval('public.cleaner_available_now_id_seq'::regclass);

--
-- Name: cleaner_profiles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_profiles ALTER COLUMN id SET DEFAULT nextval('public.cleaner_profiles_id_seq'::regclass);

--
-- Name: cleaner_reports id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_reports ALTER COLUMN id SET DEFAULT nextval('public.cleaner_reports_id_seq'::regclass);

--
-- Name: client_cleaner_notes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_cleaner_notes ALTER COLUMN id SET DEFAULT nextval('public.client_cleaner_notes_id_seq'::regclass);

--
-- Name: companies id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.companies ALTER COLUMN id SET DEFAULT nextval('public.companies_id_seq'::regclass);

--
-- Name: company_members id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_members ALTER COLUMN id SET DEFAULT nextval('public.company_members_id_seq'::regclass);

--
-- Name: conversations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations ALTER COLUMN id SET DEFAULT nextval('public.conversations_id_seq'::regclass);

--
-- Name: favorite_cleaners id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorite_cleaners ALTER COLUMN id SET DEFAULT nextval('public.favorite_cleaners_id_seq'::regclass);

--
-- Name: job_alerts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_alerts ALTER COLUMN id SET DEFAULT nextval('public.job_alerts_id_seq'::regclass);

--
-- Name: job_invitations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_invitations ALTER COLUMN id SET DEFAULT nextval('public.job_invitations_id_seq'::regclass);

--
-- Name: jobs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs ALTER COLUMN id SET DEFAULT nextval('public.jobs_id_seq'::regclass);

--
-- Name: messages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages ALTER COLUMN id SET DEFAULT nextval('public.messages_id_seq1'::regclass);

--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);

--
-- Name: platform_access_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_access_settings ALTER COLUMN id SET DEFAULT nextval('public.platform_access_settings_id_seq'::regclass);

--
-- Name: preferred_cleaners id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.preferred_cleaners ALTER COLUMN id SET DEFAULT nextval('public.preferred_cleaners_id_seq'::regclass);

--
-- Name: profile_media id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_media ALTER COLUMN id SET DEFAULT nextval('public.profile_media_id_seq'::regclass);

--
-- Name: promo_codes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promo_codes ALTER COLUMN id SET DEFAULT nextval('public.promo_codes_id_seq'::regclass);

--
-- Name: recently_viewed_cleaners id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recently_viewed_cleaners ALTER COLUMN id SET DEFAULT nextval('public.recently_viewed_cleaners_id_seq'::regclass);

--
-- Name: recurring_availability id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recurring_availability ALTER COLUMN id SET DEFAULT nextval('public.recurring_availability_id_seq'::regclass);

--
-- Name: referral_codes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_codes ALTER COLUMN id SET DEFAULT nextval('public.referral_codes_id_seq'::regclass);

--
-- Name: referral_redemptions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_redemptions ALTER COLUMN id SET DEFAULT nextval('public.referral_redemptions_id_seq'::regclass);

--
-- Name: repeat_booking_requests id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_booking_requests ALTER COLUMN id SET DEFAULT nextval('public.repeat_booking_requests_id_seq'::regclass);

--
-- Name: repeat_bookings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_bookings ALTER COLUMN id SET DEFAULT nextval('public.repeat_bookings_id_seq'::regclass);

--
-- Name: reports id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports ALTER COLUMN id SET DEFAULT nextval('public.reports_id_seq'::regclass);

--
-- Name: reviews id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews ALTER COLUMN id SET DEFAULT nextval('public.reviews_id_seq'::regclass);

--
-- Name: saved_jobs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_jobs ALTER COLUMN id SET DEFAULT nextval('public.saved_jobs_id_seq'::regclass);

--
-- Name: subscription_plans id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscription_plans ALTER COLUMN id SET DEFAULT nextval('public.subscription_plans_id_seq'::regclass);

--
-- Name: user_daily_access id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_daily_access ALTER COLUMN id SET DEFAULT nextval('public.user_daily_access_id_seq'::regclass);

--
-- Name: user_subscriptions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_subscriptions ALTER COLUMN id SET DEFAULT nextval('public.user_subscriptions_id_seq'::regclass);

--
-- Name: user_usage id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_usage ALTER COLUMN id SET DEFAULT nextval('public.user_usage_id_seq'::regclass);

--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);

--
-- Name: verification_requests id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_requests ALTER COLUMN id SET DEFAULT nextval('public.verification_requests_id_seq'::regclass);

--
-- Name: work_proofs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_proofs ALTER COLUMN id SET DEFAULT nextval('public.work_proofs_id_seq'::regclass);

--
-- Name: application_timeline application_timeline_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.application_timeline
    ADD CONSTRAINT application_timeline_pkey PRIMARY KEY (id);

--
-- Name: applications applications_job_id_cleaner_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applications
    ADD CONSTRAINT applications_job_id_cleaner_id_key UNIQUE (job_id, cleaner_id);

--
-- Name: applications applications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applications
    ADD CONSTRAINT applications_pkey PRIMARY KEY (id);

--
-- Name: availability_overrides availability_overrides_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.availability_overrides
    ADD CONSTRAINT availability_overrides_pkey PRIMARY KEY (id);

--
-- Name: availability_settings availability_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.availability_settings
    ADD CONSTRAINT availability_settings_pkey PRIMARY KEY (cleaner_id);

--
-- Name: billing_checkout_sessions billing_checkout_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_checkout_sessions
    ADD CONSTRAINT billing_checkout_sessions_pkey PRIMARY KEY (id);

--
-- Name: billing_checkout_sessions billing_checkout_sessions_stripe_session_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_checkout_sessions
    ADD CONSTRAINT billing_checkout_sessions_stripe_session_id_key UNIQUE (stripe_session_id);

--
-- Name: billing_customers billing_customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_customers
    ADD CONSTRAINT billing_customers_pkey PRIMARY KEY (id);

--
-- Name: billing_customers billing_customers_stripe_customer_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_customers
    ADD CONSTRAINT billing_customers_stripe_customer_id_key UNIQUE (stripe_customer_id);

--
-- Name: billing_customers billing_customers_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_customers
    ADD CONSTRAINT billing_customers_user_id_key UNIQUE (user_id);

--
-- Name: blocked_cleaners blocked_cleaners_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocked_cleaners
    ADD CONSTRAINT blocked_cleaners_pkey PRIMARY KEY (id);

--
-- Name: booking_status_history booking_status_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.booking_status_history
    ADD CONSTRAINT booking_status_history_pkey PRIMARY KEY (id);

--
-- Name: bookings bookings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_pkey PRIMARY KEY (id);

--
-- Name: cleaner_availability cleaner_availability_cleaner_id_available_date_start_time_e_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_availability
    ADD CONSTRAINT cleaner_availability_cleaner_id_available_date_start_time_e_key UNIQUE (cleaner_id, available_date, start_time, end_time);

--
-- Name: cleaner_availability cleaner_availability_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_availability
    ADD CONSTRAINT cleaner_availability_pkey PRIMARY KEY (id);

--
-- Name: cleaner_available_now cleaner_available_now_cleaner_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_available_now
    ADD CONSTRAINT cleaner_available_now_cleaner_id_key UNIQUE (cleaner_id);

--
-- Name: cleaner_available_now cleaner_available_now_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_available_now
    ADD CONSTRAINT cleaner_available_now_pkey PRIMARY KEY (id);

--
-- Name: cleaner_profiles cleaner_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_profiles
    ADD CONSTRAINT cleaner_profiles_pkey PRIMARY KEY (id);

--
-- Name: cleaner_profiles cleaner_profiles_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_profiles
    ADD CONSTRAINT cleaner_profiles_user_id_key UNIQUE (user_id);

--
-- Name: cleaner_reports cleaner_reports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_reports
    ADD CONSTRAINT cleaner_reports_pkey PRIMARY KEY (id);

--
-- Name: cleaner_reputation cleaner_reputation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_reputation
    ADD CONSTRAINT cleaner_reputation_pkey PRIMARY KEY (cleaner_id);

--
-- Name: client_cleaner_notes client_cleaner_notes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_cleaner_notes
    ADD CONSTRAINT client_cleaner_notes_pkey PRIMARY KEY (id);

--
-- Name: companies companies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.companies
    ADD CONSTRAINT companies_pkey PRIMARY KEY (id);

--
-- Name: company_members company_members_company_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_members
    ADD CONSTRAINT company_members_company_id_user_id_key UNIQUE (company_id, user_id);

--
-- Name: company_members company_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_members
    ADD CONSTRAINT company_members_pkey PRIMARY KEY (id);

--
-- Name: conversations conversations_booking_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_booking_id_key UNIQUE (booking_id);

--
-- Name: conversations conversations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_pkey PRIMARY KEY (id);

--
-- Name: favorite_cleaners favorite_cleaners_client_id_cleaner_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorite_cleaners
    ADD CONSTRAINT favorite_cleaners_client_id_cleaner_id_key UNIQUE (client_id, cleaner_id);

--
-- Name: favorite_cleaners favorite_cleaners_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorite_cleaners
    ADD CONSTRAINT favorite_cleaners_pkey PRIMARY KEY (id);

--
-- Name: job_alerts job_alerts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_alerts
    ADD CONSTRAINT job_alerts_pkey PRIMARY KEY (id);

--
-- Name: job_invitations job_invitations_job_id_cleaner_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_invitations
    ADD CONSTRAINT job_invitations_job_id_cleaner_id_key UNIQUE (job_id, cleaner_id);

--
-- Name: job_invitations job_invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_invitations
    ADD CONSTRAINT job_invitations_pkey PRIMARY KEY (id);

--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (id);

--
-- Name: messages messages_pkey1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey1 PRIMARY KEY (id);

--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);

--
-- Name: platform_access_settings platform_access_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_access_settings
    ADD CONSTRAINT platform_access_settings_pkey PRIMARY KEY (id);

--
-- Name: preferred_cleaners preferred_cleaners_client_id_cleaner_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.preferred_cleaners
    ADD CONSTRAINT preferred_cleaners_client_id_cleaner_id_key UNIQUE (client_id, cleaner_id);

--
-- Name: preferred_cleaners preferred_cleaners_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.preferred_cleaners
    ADD CONSTRAINT preferred_cleaners_pkey PRIMARY KEY (id);

--
-- Name: profile_media profile_media_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_media
    ADD CONSTRAINT profile_media_pkey PRIMARY KEY (id);

--
-- Name: promo_codes promo_codes_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promo_codes
    ADD CONSTRAINT promo_codes_code_key UNIQUE (code);

--
-- Name: promo_codes promo_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promo_codes
    ADD CONSTRAINT promo_codes_pkey PRIMARY KEY (id);

--
-- Name: recently_viewed_cleaners recently_viewed_cleaners_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recently_viewed_cleaners
    ADD CONSTRAINT recently_viewed_cleaners_pkey PRIMARY KEY (id);

--
-- Name: recurring_availability recurring_availability_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recurring_availability
    ADD CONSTRAINT recurring_availability_pkey PRIMARY KEY (id);

--
-- Name: referral_codes referral_codes_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_codes
    ADD CONSTRAINT referral_codes_code_key UNIQUE (code);

--
-- Name: referral_codes referral_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_codes
    ADD CONSTRAINT referral_codes_pkey PRIMARY KEY (id);

--
-- Name: referral_redemptions referral_redemptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_redemptions
    ADD CONSTRAINT referral_redemptions_pkey PRIMARY KEY (id);

--
-- Name: referral_redemptions referral_redemptions_referral_code_id_referred_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_redemptions
    ADD CONSTRAINT referral_redemptions_referral_code_id_referred_user_id_key UNIQUE (referral_code_id, referred_user_id);

--
-- Name: repeat_booking_requests repeat_booking_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_booking_requests
    ADD CONSTRAINT repeat_booking_requests_pkey PRIMARY KEY (id);

--
-- Name: repeat_bookings repeat_bookings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_bookings
    ADD CONSTRAINT repeat_bookings_pkey PRIMARY KEY (id);

--
-- Name: reports reports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_pkey PRIMARY KEY (id);

--
-- Name: reviews reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_pkey PRIMARY KEY (id);

--
-- Name: saved_jobs saved_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_jobs
    ADD CONSTRAINT saved_jobs_pkey PRIMARY KEY (id);

--
-- Name: saved_jobs saved_jobs_user_id_job_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_jobs
    ADD CONSTRAINT saved_jobs_user_id_job_id_key UNIQUE (user_id, job_id);

--
-- Name: subscription_plans subscription_plans_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscription_plans
    ADD CONSTRAINT subscription_plans_name_key UNIQUE (name);

--
-- Name: subscription_plans subscription_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscription_plans
    ADD CONSTRAINT subscription_plans_pkey PRIMARY KEY (id);

--
-- Name: user_daily_access user_daily_access_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_daily_access
    ADD CONSTRAINT user_daily_access_pkey PRIMARY KEY (id);

--
-- Name: user_daily_access user_daily_access_user_id_access_date_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_daily_access
    ADD CONSTRAINT user_daily_access_user_id_access_date_key UNIQUE (user_id, access_date);

--
-- Name: user_subscriptions user_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_pkey PRIMARY KEY (id);

--
-- Name: user_subscriptions user_subscriptions_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_user_id_key UNIQUE (user_id);

--
-- Name: user_usage user_usage_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_usage
    ADD CONSTRAINT user_usage_pkey PRIMARY KEY (id);

--
-- Name: user_usage user_usage_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_usage
    ADD CONSTRAINT user_usage_user_id_key UNIQUE (user_id);

--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

--
-- Name: verification_requests verification_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_requests
    ADD CONSTRAINT verification_requests_pkey PRIMARY KEY (id);

--
-- Name: work_proofs work_proofs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_proofs
    ADD CONSTRAINT work_proofs_pkey PRIMARY KEY (id);

--
-- Name: booking_status_history_booking_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX booking_status_history_booking_id_idx ON public.booking_status_history USING btree (booking_id);

--
-- Name: booking_status_history_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX booking_status_history_created_at_idx ON public.booking_status_history USING btree (created_at);

--
-- Name: conversations_cleaner_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX conversations_cleaner_id_idx ON public.conversations USING btree (cleaner_id);

--
-- Name: conversations_client_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX conversations_client_id_idx ON public.conversations USING btree (client_id);

--
-- Name: conversations_last_message_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX conversations_last_message_at_idx ON public.conversations USING btree (last_message_at DESC);

--
-- Name: idx_application_timeline_application; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_application_timeline_application ON public.application_timeline USING btree (application_id);

--
-- Name: idx_application_timeline_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_application_timeline_created_at ON public.application_timeline USING btree (application_id, created_at);

--
-- Name: idx_availability_overrides_cleaner_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_availability_overrides_cleaner_date ON public.availability_overrides USING btree (cleaner_id, available_date);

--
-- Name: idx_bookings_cleaner_schedule; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_bookings_cleaner_schedule ON public.bookings USING btree (cleaner_id, scheduled_at, scheduled_end_at) WHERE (status = ANY (ARRAY['pending'::text, 'confirmed'::text, 'in_progress'::text]));

--
-- Name: idx_cleaner_available_now_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cleaner_available_now_active ON public.cleaner_available_now USING btree (is_available, available_until);

--
-- Name: idx_cleaner_available_now_job_types; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cleaner_available_now_job_types ON public.cleaner_available_now USING gin (job_types);

--
-- Name: idx_cleaner_available_now_location; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cleaner_available_now_location ON public.cleaner_available_now USING btree (lower(location));

--
-- Name: idx_company_members_active_cleaners; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_company_members_active_cleaners ON public.company_members USING btree (company_id, role, status);

--
-- Name: idx_profile_media_company; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_profile_media_company ON public.profile_media USING btree (company_id);

--
-- Name: idx_profile_media_owner_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_profile_media_owner_user ON public.profile_media USING btree (owner_user_id);

--
-- Name: idx_profile_media_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_profile_media_type ON public.profile_media USING btree (media_type);

--
-- Name: idx_recurring_availability_cleaner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_recurring_availability_cleaner ON public.recurring_availability USING btree (cleaner_id, weekday);

--
-- Name: idx_work_proofs_booking_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_work_proofs_booking_id ON public.work_proofs USING btree (booking_id);

--
-- Name: idx_work_proofs_booking_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_work_proofs_booking_type ON public.work_proofs USING btree (booking_id, proof_type);

--
-- Name: idx_work_proofs_cleaner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_work_proofs_cleaner_id ON public.work_proofs USING btree (cleaner_id);

--
-- Name: messages_conversation_created_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX messages_conversation_created_idx ON public.messages USING btree (conversation_id, created_at);

--
-- Name: messages_conversation_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX messages_conversation_id_idx ON public.messages USING btree (conversation_id);

--
-- Name: reviews_booking_id_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX reviews_booking_id_unique ON public.reviews USING btree (booking_id) WHERE (booking_id IS NOT NULL);

--
-- Name: application_timeline application_timeline_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.application_timeline
    ADD CONSTRAINT application_timeline_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON DELETE SET NULL;

--
-- Name: application_timeline application_timeline_application_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.application_timeline
    ADD CONSTRAINT application_timeline_application_id_fkey FOREIGN KEY (application_id) REFERENCES public.applications(id) ON DELETE CASCADE;

--
-- Name: applications applications_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applications
    ADD CONSTRAINT applications_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: applications applications_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applications
    ADD CONSTRAINT applications_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(id) ON DELETE CASCADE;

--
-- Name: availability_overrides availability_overrides_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.availability_overrides
    ADD CONSTRAINT availability_overrides_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: availability_settings availability_settings_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.availability_settings
    ADD CONSTRAINT availability_settings_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: billing_checkout_sessions billing_checkout_sessions_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_checkout_sessions
    ADD CONSTRAINT billing_checkout_sessions_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.subscription_plans(id) ON DELETE SET NULL;

--
-- Name: billing_checkout_sessions billing_checkout_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_checkout_sessions
    ADD CONSTRAINT billing_checkout_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: billing_customers billing_customers_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_customers
    ADD CONSTRAINT billing_customers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: booking_status_history booking_status_history_booking_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.booking_status_history
    ADD CONSTRAINT booking_status_history_booking_id_fkey FOREIGN KEY (booking_id) REFERENCES public.bookings(id) ON DELETE CASCADE;

--
-- Name: booking_status_history booking_status_history_changed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.booking_status_history
    ADD CONSTRAINT booking_status_history_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES public.users(id) ON DELETE SET NULL;

--
-- Name: bookings bookings_application_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_application_id_fkey FOREIGN KEY (application_id) REFERENCES public.applications(id) ON DELETE SET NULL;

--
-- Name: bookings bookings_cancelled_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_cancelled_by_fkey FOREIGN KEY (cancelled_by) REFERENCES public.users(id);

--
-- Name: bookings bookings_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: bookings bookings_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: bookings bookings_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(id) ON DELETE CASCADE;

--
-- Name: cleaner_availability cleaner_availability_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_availability
    ADD CONSTRAINT cleaner_availability_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: cleaner_available_now cleaner_available_now_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_available_now
    ADD CONSTRAINT cleaner_available_now_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: cleaner_profiles cleaner_profiles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_profiles
    ADD CONSTRAINT cleaner_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: cleaner_reports cleaner_reports_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_reports
    ADD CONSTRAINT cleaner_reports_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: cleaner_reports cleaner_reports_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_reports
    ADD CONSTRAINT cleaner_reports_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: cleaner_reputation cleaner_reputation_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cleaner_reputation
    ADD CONSTRAINT cleaner_reputation_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: companies companies_owner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.companies
    ADD CONSTRAINT companies_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: company_members company_members_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_members
    ADD CONSTRAINT company_members_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(id) ON DELETE CASCADE;

--
-- Name: company_members company_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_members
    ADD CONSTRAINT company_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: conversations conversations_booking_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_booking_fk FOREIGN KEY (booking_id) REFERENCES public.bookings(id) ON DELETE CASCADE;

--
-- Name: conversations conversations_cleaner_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_cleaner_fk FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: conversations conversations_client_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_client_fk FOREIGN KEY (client_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: favorite_cleaners favorite_cleaners_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorite_cleaners
    ADD CONSTRAINT favorite_cleaners_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: favorite_cleaners favorite_cleaners_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorite_cleaners
    ADD CONSTRAINT favorite_cleaners_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: job_alerts job_alerts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_alerts
    ADD CONSTRAINT job_alerts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: job_invitations job_invitations_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_invitations
    ADD CONSTRAINT job_invitations_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: job_invitations job_invitations_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_invitations
    ADD CONSTRAINT job_invitations_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: job_invitations job_invitations_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_invitations
    ADD CONSTRAINT job_invitations_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(id) ON DELETE CASCADE;

--
-- Name: jobs jobs_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: messages messages_conversation_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_conversation_fk FOREIGN KEY (conversation_id) REFERENCES public.conversations(id) ON DELETE CASCADE;

--
-- Name: messages messages_sender_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_sender_fk FOREIGN KEY (sender_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: preferred_cleaners preferred_cleaners_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.preferred_cleaners
    ADD CONSTRAINT preferred_cleaners_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: preferred_cleaners preferred_cleaners_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.preferred_cleaners
    ADD CONSTRAINT preferred_cleaners_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: profile_media profile_media_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_media
    ADD CONSTRAINT profile_media_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(id) ON DELETE CASCADE;

--
-- Name: profile_media profile_media_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_media
    ADD CONSTRAINT profile_media_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: promo_codes promo_codes_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promo_codes
    ADD CONSTRAINT promo_codes_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.subscription_plans(id) ON DELETE SET NULL;

--
-- Name: recurring_availability recurring_availability_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recurring_availability
    ADD CONSTRAINT recurring_availability_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: referral_codes referral_codes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_codes
    ADD CONSTRAINT referral_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: referral_redemptions referral_redemptions_referral_code_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_redemptions
    ADD CONSTRAINT referral_redemptions_referral_code_id_fkey FOREIGN KEY (referral_code_id) REFERENCES public.referral_codes(id) ON DELETE CASCADE;

--
-- Name: referral_redemptions referral_redemptions_referred_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_redemptions
    ADD CONSTRAINT referral_redemptions_referred_user_id_fkey FOREIGN KEY (referred_user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: referral_redemptions referral_redemptions_referrer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_redemptions
    ADD CONSTRAINT referral_redemptions_referrer_id_fkey FOREIGN KEY (referrer_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: repeat_booking_requests repeat_booking_requests_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_booking_requests
    ADD CONSTRAINT repeat_booking_requests_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: repeat_booking_requests repeat_booking_requests_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_booking_requests
    ADD CONSTRAINT repeat_booking_requests_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: repeat_booking_requests repeat_booking_requests_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_booking_requests
    ADD CONSTRAINT repeat_booking_requests_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(id) ON DELETE CASCADE;

--
-- Name: repeat_booking_requests repeat_booking_requests_new_booking_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_booking_requests
    ADD CONSTRAINT repeat_booking_requests_new_booking_id_fkey FOREIGN KEY (new_booking_id) REFERENCES public.bookings(id) ON DELETE SET NULL;

--
-- Name: repeat_booking_requests repeat_booking_requests_original_booking_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_booking_requests
    ADD CONSTRAINT repeat_booking_requests_original_booking_id_fkey FOREIGN KEY (original_booking_id) REFERENCES public.bookings(id) ON DELETE CASCADE;

--
-- Name: repeat_bookings repeat_bookings_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_bookings
    ADD CONSTRAINT repeat_bookings_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: repeat_bookings repeat_bookings_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_bookings
    ADD CONSTRAINT repeat_bookings_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: repeat_bookings repeat_bookings_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repeat_bookings
    ADD CONSTRAINT repeat_bookings_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(id) ON DELETE CASCADE;

--
-- Name: reports reports_booking_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_booking_id_fkey FOREIGN KEY (booking_id) REFERENCES public.bookings(id) ON DELETE SET NULL;

--
-- Name: reports reports_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(id) ON DELETE SET NULL;

--
-- Name: reports reports_reported_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_reported_user_id_fkey FOREIGN KEY (reported_user_id) REFERENCES public.users(id) ON DELETE SET NULL;

--
-- Name: reports reports_reporter_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_reporter_id_fkey FOREIGN KEY (reporter_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: reports reports_reviewed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.users(id) ON DELETE SET NULL;

--
-- Name: reviews reviews_booking_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_booking_id_fkey FOREIGN KEY (booking_id) REFERENCES public.bookings(id) ON DELETE CASCADE;

--
-- Name: reviews reviews_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: reviews reviews_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: reviews reviews_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(id) ON DELETE CASCADE;

--
-- Name: saved_jobs saved_jobs_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_jobs
    ADD CONSTRAINT saved_jobs_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(id) ON DELETE CASCADE;

--
-- Name: saved_jobs saved_jobs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_jobs
    ADD CONSTRAINT saved_jobs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: user_daily_access user_daily_access_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_daily_access
    ADD CONSTRAINT user_daily_access_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: user_subscriptions user_subscriptions_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.subscription_plans(id) ON DELETE SET NULL;

--
-- Name: user_subscriptions user_subscriptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: user_usage user_usage_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_usage
    ADD CONSTRAINT user_usage_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: verification_requests verification_requests_reviewed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_requests
    ADD CONSTRAINT verification_requests_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.users(id) ON DELETE SET NULL;

--
-- Name: verification_requests verification_requests_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_requests
    ADD CONSTRAINT verification_requests_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: work_proofs work_proofs_booking_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_proofs
    ADD CONSTRAINT work_proofs_booking_id_fkey FOREIGN KEY (booking_id) REFERENCES public.bookings(id) ON DELETE CASCADE;

--
-- Name: work_proofs work_proofs_cleaner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_proofs
    ADD CONSTRAINT work_proofs_cleaner_id_fkey FOREIGN KEY (cleaner_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--
