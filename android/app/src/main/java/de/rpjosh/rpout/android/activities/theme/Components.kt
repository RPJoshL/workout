package de.rpjosh.rpout.android.activities.theme

import androidx.compose.animation.Animatable
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.LocalIndication
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.wrapContentWidth
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.ArrowDropDown
import androidx.compose.material.ripple.rememberRipple
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedCard
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.Modifier
import androidx.compose.ui.autofill.AutofillNode
import androidx.compose.ui.autofill.AutofillType
import androidx.compose.ui.composed
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.focus.onFocusEvent
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.layout.boundsInWindow
import androidx.compose.ui.layout.onGloballyPositioned
import androidx.compose.ui.layout.positionInParent
import androidx.compose.ui.layout.positionInRoot
import androidx.compose.ui.platform.LocalAutofill
import androidx.compose.ui.platform.LocalAutofillTree
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.OffsetMapping
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.TransformedText
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.text.style.LineHeightStyle
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.unit.toSize
import androidx.compose.ui.window.Popup
import kotlin.math.exp
import kotlin.math.roundToInt

/**
 * Renders a small and compact text box
 */
@Composable
fun OutlinedTextField(placeholder: String? = null, value: String, onValueChange: (String) -> Unit, password: Boolean = false, autoCorrect: Boolean = false, modifier: Modifier? = null) {
    val borderColor = remember { Animatable(textDarker) }
    var hasFocus by remember { mutableStateOf(false) }

    var visualTransformation: VisualTransformation = VisualTransformation.None
    val keyboardOptions: KeyboardOptions

    if (password) {
        visualTransformation = PasswordVisualTransformation()
        keyboardOptions = KeyboardOptions( keyboardType = KeyboardType.Password, autoCorrectEnabled = autoCorrect )
    } else {
        keyboardOptions = KeyboardOptions( autoCorrectEnabled = autoCorrect )
    }

    LaunchedEffect(hasFocus) {
        if (hasFocus) borderColor.animateTo(accentBlueBorder, animationSpec = tween(300))
        else borderColor.animateTo(textDarker, animationSpec = tween(300))
    }

    // Set default modifier
    val modifierO = modifier ?: Modifier.fillMaxWidth()

    BasicTextField(
        modifier = modifierO.onFocusEvent {
            hasFocus = it.hasFocus
        },
        value = value,
        onValueChange = onValueChange,
        textStyle = TextStyle(
            color = text,
            fontSize = 15.sp,
            letterSpacing = 0.6.sp
        ),
        cursorBrush = SolidColor(accentBlueBorder),
        visualTransformation = visualTransformation,
        keyboardOptions = keyboardOptions,

        decorationBox = { innerTextField ->
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(
                        color = backgroundDisabledDarker,
                        shape = RoundedCornerShape(5.dp)
                    )
                    .border(
                        width = if (hasFocus) 2.dp else 1.dp,
                        color =  borderColor.value,
                        shape = RoundedCornerShape(5.dp)
                    )
                    .padding(11.dp)
            ) {
                Box() {
                    if (value.isEmpty() && placeholder != null) {
                        Text(
                            text = placeholder,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis,
                            color = textHint,
                            fontSize = 15.sp,
                            modifier = Modifier.padding(0.dp),
                            lineHeight = 3.sp
                        )
                    }
                    innerTextField()
                }
            }

        }
    )
}

/**
 * Renders a (bigger) outlined text box with a label on top when the user has set a value
 */
@Composable
fun OutlinedTextFieldWithLabel(label: String? = null, placeholder: String? = null, value: String, onValueChange: (String) -> Unit, password: Boolean = false, autoCorrect: Boolean = false, modifier: Modifier? = null) {

    var visualTransformation: VisualTransformation = VisualTransformation.None
    val keyboardOptions: KeyboardOptions

    if (password) {
        visualTransformation = PasswordVisualTransformation()
        keyboardOptions = KeyboardOptions( keyboardType = KeyboardType.Password, autoCorrectEnabled = autoCorrect )
    } else {
        keyboardOptions = KeyboardOptions( autoCorrectEnabled = autoCorrect )
    }

    // Always show label and placeholder
    if (placeholder != null && label != null && value.isEmpty()) {
        visualTransformation = PlaceholderTransformation(placeholder)
    }

    // Set default modifier
    val modifierO = modifier ?: Modifier.fillMaxWidth()

    androidx.compose.material3.OutlinedTextField(
        modifier = modifierO,
        value = value,
        onValueChange = onValueChange,
        label = { Text(label ?: "") },
        placeholder = { Text(placeholder ?: label ?: "") },
        maxLines = 1,
        shape = RoundedCornerShape(5.dp),
        textStyle = TextStyle(
            color = if(value.isEmpty()) textHint else text,
            fontSize = 16.sp,
            letterSpacing = 0.6.sp
        ),
        colors = OutlinedTextFieldDefaults.colors(
            focusedContainerColor = backgroundDisabledDarker,
            unfocusedContainerColor = backgroundDisabledDarker,
            focusedBorderColor = accentBlueBorder,
            unfocusedBorderColor = textDarker,
            cursorColor = accentBlueBorder,
            focusedLabelColor = text,
        ),
        visualTransformation = visualTransformation,
        keyboardOptions = keyboardOptions
    )
}

class PlaceholderTransformation(val placeholder: String) : VisualTransformation {
    override fun filter(text: AnnotatedString): TransformedText {
        return PlaceholderFilter(text, placeholder)
    }
}

fun PlaceholderFilter(text: AnnotatedString, placeholder: String): TransformedText {

    val numberOffsetTranslator = object : OffsetMapping {
        override fun originalToTransformed(offset: Int): Int {
            return 0
        }

        override fun transformedToOriginal(offset: Int): Int {
            return 0
        }
    }

    return TransformedText(AnnotatedString(placeholder), numberOffsetTranslator)
}

@OptIn(ExperimentalComposeUiApi::class)
fun Modifier.autofill(
    autofillTypes: List<AutofillType>,
    onFill: ((String) -> Unit),
) = composed {
    val autofill = LocalAutofill.current
    val autofillNode = AutofillNode(onFill = onFill, autofillTypes = autofillTypes)
    LocalAutofillTree.current += autofillNode

    this.onGloballyPositioned {
        autofillNode.boundingBox = it.boundsInWindow()
    }.onFocusChanged { focusState ->
        autofill?.run {
            if (focusState.isFocused) {
                requestAutofillForNode(autofillNode)
            } else {
                cancelAutofillForNode(autofillNode)
            }
        }
    }
}

/** Spinner to select a single value from */
@Composable
fun Spinner(
    options: List<SelectOption>,
    preselected: SelectOption,
    onSelectionChanged: (data: SelectOption) -> Unit,
    modifier: Modifier = Modifier
) {

    var selected by remember { mutableStateOf(preselected) }
    var expanded by remember { mutableStateOf(false) }
    var rowSize by remember { mutableStateOf(Size.Zero) }

    var yOffset by remember { mutableIntStateOf(0) }

    Column() {
        Card(
            modifier = modifier.clickable {
                expanded = !expanded
            }.onGloballyPositioned {
                rowSize = it.size.toSize()
            },
            colors = CardDefaults.cardColors(
                contentColor = text,
                containerColor = backgroundDarker
            ),
            border = BorderStroke(1.dp, textHint)
        ) {
            Column {
                Row(
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.Top,
                ) {

                    Text(
                        text = selected.label,
                        modifier = Modifier.weight(1f)
                            .padding(horizontal = 16.dp, vertical = 8.dp)
                    )
                    Icon(Icons.Outlined.ArrowDropDown, null, modifier = Modifier.padding(8.dp))
                }
                // Row to the the y position offset
                Row(modifier = Modifier.onGloballyPositioned { yOffset = it.positionInParent().y.toInt() }) {}
            }

        }

        // We don't use a DropdownMenu because it cannot be customized as we want
        Popup(offset = IntOffset(
                with(LocalDensity.current) { 2.dp.toPx().toInt() },
                y = yOffset + with(LocalDensity.current) { 6.dp.toPx().toInt() }
            )
        ) {
            AnimatedVisibility(
                visible = expanded,
                enter = fadeIn(animationSpec = tween(600)),
                exit = fadeOut(animationSpec = tween(600)),
            ) {
                Box(
                    modifier = modifier.background(backgroundDarker)
                        .border(1.dp, textHint, shape = RoundedCornerShape(5.dp))
                        .width(with(LocalDensity.current) { rowSize.width.toDp() - 4.dp })
                ) {
                    Column(
                        modifier = Modifier.padding(4.dp)
                    ) {
                        options.forEach { listEntry ->
                            // No TextButton because of too big padding
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clickable(
                                        onClick = {
                                            selected = listEntry
                                            expanded = false
                                            onSelectionChanged(selected)
                                        },
                                        indication = LocalIndication.current,
                                        interactionSource = remember { MutableInteractionSource() }
                                    ),
                            ) {
                                Text(
                                    textAlign = TextAlign.Start,
                                    text = listEntry.label,
                                    color = if (listEntry.id == selected.id) textBlue else text,
                                    fontSize = 14.sp,
                                    modifier = Modifier
                                        .padding(top = 4.dp, bottom = 4.dp, start = 8.dp)
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}


data class SelectOption (
    val id: Int,
    val label: String
)
