package emulator
var (
	CLOCK_REALTIME   uint64 = 0
	CLOCK_MONOTONIC  uint64 = 1
	CLOCK_PROCESS_CPUTIME_ID uint64 = 2
	CLOCK_THREAD_CPUTIME_ID  uint64 = 3
	CLOCK_MONOTONIC_RAW      uint64 = 4
	CLOCK_REALTIME_COARSE    uint64 = 5
	CLOCK_MONOTONIC_COARSE   uint64 = 6
	CLOCK_BOOTTIME       uint64 = 7
	CLOCK_REALTIME_ALARM uint64 = 8
	CLOCK_BOOTTIME_ALARM uint64 = 9

	FUTEX_WAIT uint64 = 0
	FUTEX_WAKE uint64 = 1
	FUTEX_FD uint64 = 2
	FUTEX_REQUEUE uint64 = 3
	FUTEX_CMP_REQUEUE uint64 = 4
	FUTEX_WAKE_OP uint64 = 5
	FUTEX_LOCK_PI uint64 = 6
	FUTEX_UNLOCK_PI uint64 = 7
	FUTEX_TRYLOCK_PI uint64 = 8
	FUTEX_WAIT_BITSET uint64 = 9
	FUTEX_WAKE_BITSET uint64 = 10
	FUTEX_WAIT_REQUEUE_PI uint64 = 11
	FUTEX_CMP_REQUEUE_PI uint64 = 12

	FUTEX_PRIVATE_FLAG uint64 = 128
	FUTEX_CLOCK_REALTIME uint64 = 256

	tmp int64 = ^int64(FUTEX_PRIVATE_FLAG | FUTEX_CLOCK_REALTIME)
	FUTEX_CMD_MASK uint64 = uint64(tmp)

	//fcntl
	/* command values */
	F_DUPFD  uint64=0		/* duplicate file descriptor */
	F_GETFD  uint64=1		/* get file descriptor flags */
	F_SETFD  uint64=2		/* set file descriptor flags */
	F_GETFL  uint64=3		/* get file status flags */
	F_SETFL  uint64=4		/* set file status flags */
	F_GETOWN uint64=5		/* get SIGIO/SIGURG proc/pgrp */
	F_SETOWN uint64=6		/* set SIGIO/SIGURG proc/pgrp */
	F_GETLK  uint64=7		/* get record locking information */
	F_SETLK  uint64=8		/* set record locking information */
	F_SETLKW uint64=9		/* F_SETLK; wait if blocked */

	/* file descriptor flags (F_GETFD, F_SETFD) */
	FD_CLOEXEC uint64=1		/* close-on-exec flag */

	/* record locking flags (F_GETLK, F_SETLK, F_SETLKW) */
	F_RDLCK uint64=1		/* shared or read lock */
	F_UNLCK uint64=2		/* unlock */
	F_WRLCK uint64=3		/* exclusive or write lock */
	F_WAIT  uint64=0x010		/* Wait until lock is granted */
	F_FLOCK uint64=0x020	 	/* Use flock(2) semantics for lock */
	F_POSIX uint64=0x040	 	/* Use POSIX semantics for lock */
)