#!/bin/sh
export LC_ALL="en_US.UTF-8"

PROC_KEYPAD="/proc/keypad"
PROC_FIVEWAY="/proc/fiveway"
[ -e $PROC_KEYPAD ] && echo unlock > $PROC_KEYPAD
[ -e $PROC_FIVEWAY ] && echo unlock > $PROC_FIVEWAY

# Handle logging...
logmsg()
{
	# Use the right tools for the platform
	if [ "${INIT_TYPE}" == "sysv" ] ; then
		msg "kindle-weather: ${1}" "I"
	elif [ "${INIT_TYPE}" == "upstart" ] ; then
		f_log I kindleweather wrapper "" "${1}"
	fi

	# And throw that on stdout too, for the DIY crowd ;)
	echo "${1}"
}

# Keep track of what we do with pillow...
export AWESOME_STOPPED="no"
PILLOW_HARD_DISABLED="no"
PILLOW_SOFT_DISABLED="no"

# Keep track of if we were started through KUAL
FROM_KUAL="no"

STOP_FRAMEWORK="yes"
NO_SLEEP="yes"

# Detect if we were started by KUAL by checking our nice value...
if [ "$(nice)" == "5" ] ; then
	FROM_KUAL="yes"
	if [ "${NO_SLEEP}" == "no" ] ; then
		# Yield a bit to let stuff stop properly...
		logmsg "Hush now . . ."
		# NOTE: This may or may not be terribly useful...
		usleep 250000
	fi

	# Kindlet threads spawn with a nice value of 5, go back to a neutral value
	logmsg "Be nice!"
	renice -n -5 $$
fi

# check if we are supposed to shut down the Amazon framework
if [ "${STOP_FRAMEWORK}" == "yes" ] ; then
	logmsg "Stopping the framework . . ."
	# Upstart or SysV?
	if [ "${INIT_TYPE}" == "sysv" ] ; then
		/etc/init.d/framework stop
	else
		# The framework job sends a SIGTERM on stop, trap it so we don't get killed if we were launched by KUAL
		trap "" SIGTERM
		stop lab126_gui
		# NOTE: Let the framework teardown finish, so we don't start before the black screen...
		usleep 1250000
		# And remove the trap like a ninja now!
		trap - SIGTERM
	fi
fi

# check if kpvbooklet was launched for more than once, if not we will disable pillow
# there's no pillow if we stopped the framework, and it's only there on systems with upstart anyway
if [ "${STOP_FRAMEWORK}" == "no" -a "${INIT_TYPE}" == "upstart" ] ; then
	count=$(lipc-get-prop -eiq com.github.kindleweather.timer count)
	if [ "$count" == "" -o "$count" == "0" ] ; then
		# NOTE: Dump the fb so we can restore something useful on exit...
		cat /dev/fb0 > /var/tmp/kindleweather-fb.dump
		# NOTE: We want to disable the status bar (at the very least). Unfortunately, the soft hide/unhide method doesn't work properly anymore since FW 5.6.5...
		if [ "$(printf "%.3s" $(grep '^Kindle 5' /etc/prettyversion.txt 2>&1 | sed -n -r 's/^(Kindle)([[:blank:]]*)([[:digit:].]*)(.*?)$/\3/p' | tr -d '.'))" -ge "565" ] ; then
			PILLOW_HARD_DISABLED="yes"
			# FIXME: So we resort to killing pillow completely on FW >= 5.6.5...
			logmsg "Disabling pillow . . ."
			lipc-set-prop com.lab126.pillow disableEnablePillow disable
			# NOTE: And, oh, joy, on FW >= 5.7.2, this is not enough to prevent the clock from refreshing, so, take the bull by the horns, and SIGSTOP the WM while we run...
			if [ "$(printf "%.3s" $(grep '^Kindle 5' /etc/prettyversion.txt 2>&1 | sed -n -r 's/^(Kindle)([[:blank:]]*)([[:digit:].]*)(.*?)$/\3/p' | tr -d '.'))" -ge "572" ] ; then
				logmsg "Stopping awesome . . ."
				killall -stop awesome
				AWESOME_STOPPED="yes"
			fi
		else
			logmsg "Hiding the status bar . . ."
			# NOTE: One more great find from eureka (http://www.mobileread.com/forums/showpost.php?p=2454141&postcount=34)
			lipc-set-prop com.lab126.pillow interrogatePillow '{"pillowId": "default_status_bar", "function": "nativeBridge.hideMe();"}'
			PILLOW_SOFT_DISABLED="yes"
		fi
		# NOTE: We don't need to sleep at all if we've already SIGSTOPped awesome ;)
		if [ "${NO_SLEEP}" == "no" -a "${AWESOME_STOPPED}" == "no" ] ; then
			# NOTE: Leave the framework time to refresh the screen, so we don't start before it has finished redrawing after collapsing the title bar
			usleep 250000
			# NOTE: If we were started from KUAL, we risk getting a list item to popup right over us, so, wait some more...
			# The culprit appears to be a I WindowManager:flashTimeoutExpired:window=Root 0 0 600x30
			if [ "${FROM_KUAL}" == "yes" ] ; then
				logmsg "Playing possum to wait for the window manager . . ."
				usleep 2500000
			fi
		fi
	fi
fi

# stop cvm (sysv & framework up only)
if [ "${STOP_FRAMEWORK}" == "no" -a "${INIT_TYPE}" == "sysv" ] ; then
	logmsg "Stopping cvm . . ."
	killall -stop cvm
fi

# finally call reader
# That's not necessary when using KPVBooklet ;).
if [ "${FROM_KUAL}" == "yes" ] ; then
	eips_print_bottom_centered "Starting Kindle Weather . . ." 1
fi
./bin/kindle-weather > crash.log 2>&1

# clean up our own process tree in case the reader crashed (if needed, to avoid flooding KUAL's log)
if pidof reader.lua > /dev/null 2>&1 ; then
	logmsg "Sending a SIGTERM to stray kindle weather processes . . ."
	killall -TERM kindle-weather
fi

# Resume cvm (only if we stopped it)
if [ "${STOP_FRAMEWORK}" == "no" -a "${INIT_TYPE}" == "sysv" ] ; then
	logmsg "Resuming cvm . . ."
	killall -cont cvm
	# We need to handle the screen refresh ourselves, frontend/device/kindle/device.lua's Kindle3.exit is called before we resume cvm ;).
	echo 'send 139' > /proc/keypad
	echo 'send 139' > /proc/keypad
fi

# Restart framework (if need be)
if [ "${STOP_FRAMEWORK}" == "yes" ] ; then
	logmsg "Restarting framework . . ."
	if [ "${INIT_TYPE}" == "sysv" ] ; then
		cd / && env -u LD_LIBRARY_PATH /etc/init.d/framework start
	else
		cd / && env -u LD_LIBRARY_PATH start lab126_gui
	fi
fi

# Display chrome bar if need be (upstart & framework up only)
if [ "${STOP_FRAMEWORK}" == "no" -a "${INIT_TYPE}" == "upstart" ] ; then
	# Depending on the FW version, we may have handled things in a few different manners...
	if [ "${AWESOME_STOPPED}" == "yes" ] ; then
		logmsg "Resuming awesome . . ."
		killall -cont awesome
	fi
	if [ "${PILLOW_HARD_DISABLED}" == "yes" ] ; then
		logmsg "Enabling pillow . . ."
		lipc-set-prop com.lab126.pillow disableEnablePillow enable
		# NOTE: Try to leave the user with a slightly more useful FB content than our own last screen...
		cat /var/tmp/kindleweather-fb.dump > /dev/fb0
		rm -f /var/tmp/kindleweather-fb.dump
		lipc-set-prop com.lab126.appmgrd start app://com.lab126.booklet.home
		# NOTE: In case we ever need an extra full flash refresh...
		#eips -s w=${SCREEN_X_RES},h=${SCREEN_Y_RES} -f
	fi
	if [ "${PILLOW_SOFT_DISABLED}" == "yes" ] ; then
		logmsg "Restoring the status bar . . ."
		# NOTE: Try to leave the user with a slightly more useful FB content than our own last screen...
		cat /var/tmp/kindleweather-fb.dump > /dev/fb0
		rm -f /var/tmp/kindleweather-fb.dump
		lipc-set-prop com.lab126.pillow interrogatePillow '{"pillowId": "default_status_bar", "function": "nativeBridge.showMe();"}'
		lipc-set-prop com.lab126.appmgrd start app://com.lab126.booklet.home
	fi
fi