#!/usr/bin/env python3
import argparse
import ipaddress
import datetime
from pathlib import Path
from cryptography import x509
from cryptography.x509.oid import NameOID
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import ec
from cryptography.hazmat.backends import default_backend


def main():
    parser = argparse.ArgumentParser(
        description="Generate a self-signed TLS certificate for an IP."
    )
    parser.add_argument("ip", help="IP address to include in the certificate SAN")
    parser.add_argument(
        "--days", type=int, default=825, help="Validity in days (default: 825)"
    )
    parser.add_argument(
        "--crt",
        default="server.crt",
        help="Output certificate file (default: server.crt)",
    )
    parser.add_argument(
        "--key",
        default="server.key",
        help="Output private key file (default: server.key)",
    )
    args = parser.parse_args()

    ip = args.ip
    days = args.days
    crt_file = Path(args.crt)
    key_file = Path(args.key)

    # --- Generate ECDSA private key (P-256) ---
    key = ec.generate_private_key(ec.SECP256R1(), default_backend())

    # --- Certificate subject/issuer ---
    subject = issuer = x509.Name(
        [
            x509.NameAttribute(NameOID.COMMON_NAME, ip),
        ]
    )

    # --- SAN extension (include IP and localhost) ---
    alt_names = [
        x509.IPAddress(ipaddress.ip_address(ip)),
        x509.DNSName("localhost"),
    ]
    san = x509.SubjectAlternativeName(alt_names)

    # --- Build the certificate ---
    cert = (
        x509.CertificateBuilder()
        .subject_name(subject)
        .issuer_name(issuer)
        .public_key(key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(datetime.datetime.now(datetime.timezone.utc) - datetime.timedelta(days=1))
        .not_valid_after(datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(days=days))
        .add_extension(san, critical=False)
        .add_extension(x509.BasicConstraints(ca=False, path_length=None), critical=True)
        .add_extension(
            x509.KeyUsage(
                digital_signature=True,
                key_encipherment=True,
                content_commitment=False,
                data_encipherment=False,
                key_agreement=True,
                key_cert_sign=False,
                crl_sign=False,
                encipher_only=False,
                decipher_only=False,
            ),
            critical=True,
        )
        .add_extension(
            x509.ExtendedKeyUsage([x509.oid.ExtendedKeyUsageOID.SERVER_AUTH]),
            critical=False,
        )
        .sign(private_key=key, algorithm=hashes.SHA256(), backend=default_backend())
    )

    # --- Write to files ---
    crt_file.write_bytes(cert.public_bytes(serialization.Encoding.PEM))
    key_file.write_bytes(
        key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.PKCS8,
            encryption_algorithm=serialization.NoEncryption(),
        )
    )


if __name__ == "__main__":
    main()

